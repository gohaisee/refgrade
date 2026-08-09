package check

import (
	"context"
	"go/ast"
	"go/token"
	"strings"

	"github.com/gohaisee/refgrade/internal/astutil"
)

const (
	kafkaImport = "github.com/segmentio/kafka-go"
	amqpImport  = "github.com/rabbitmq/amqp091-go"
	natsImport  = "github.com/nats-io/nats.go"
)

var (
	mqGates = []string{"kafka-go", "rabbitmq", "nats"}

	mqAckMethods = map[string]struct{}{
		"Ack":            {},
		"AckSync":        {},
		"AckMsg":         {},
		"CommitMessages": {},
	}

	mqConsumeHints = []string{
		"ReadMessage", "FetchMessage", "Consume", "Subscribe", "QueueSubscribe",
		"AddHandler", "ConsumeClaim", "BindSync", "PullSubscribe",
	}

	mqIdempotencyHints = []string{
		"idempotent", "Idempotent", "dedup", "Dedup", "IdempotencyKey", "idempotency",
	}

	mqReconnectHints = []string{
		"Reconnect", "MaxReconnects", "ReconnectWait", "Backoff", "backoff",
		"Retry", "retry", "reconnect",
	}

	mqConnectHints = []string{
		"amqp.Dial", "nats.Connect", "kafka.Dial", "DialLeader",
	}

	mqPublishMethods = map[string]struct{}{
		"Publish":        {},
		"PublishMsg":     {},
		"WriteMessages":  {},
		"Write":          {},
	}

	mqOrderHints = []string{
		"order", "ordering", "partition key", "PartitionKey", "ordered",
	}

	natsCriticalHints = []string{
		"critical", "payment", "order", "money", "must not lose", "must-not-lose",
	}

	jetStreamHints = []string{
		"JetStream", "jetstream", "js.", "PullSubscribe", "BindSync", "Consume",
	}

	natsAckHints = []string{
		"Ack", "AckSync", "Nak", "Term", "InProgress",
	}
)

func mqInspect(mod ModuleView) (*astutil.Pool, error) {
	return poolFor(mod, astutil.Filter{SkipTestFiles: true})
}

func mqPool(mod ModuleView) (*astutil.Pool, error) {
	pool, err := mqInspect(mod)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

// consumer ack before successful processing
type Mq01 struct{ Base }

func NewMq01() *Mq01 {
	return &Mq01{Base: Base{meta: Meta{ID: "mq-01", Gates: mqGates, Domain: "messaging", DefaultSeverity: SeverityFail}}}
}

func (c *Mq01) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := mqPool(mod)
	if err != nil {
		return nil, err
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityFail)
	return scanEarlyMQAck(pool, c.ID(), sev), nil
}

// consumer handler without idempotency heuristic
type Mq02 struct{ Base }

func NewMq02() *Mq02 {
	return &Mq02{Base: Base{meta: Meta{ID: "mq-02", Gates: mqGates, Domain: "messaging", DefaultSeverity: SeverityWarn}}}
}

func (c *Mq02) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := mqPool(mod)
	if err != nil {
		return nil, err
	}
	if !moduleHasHint(pool, mqConsumeHints) {
		return nil, nil
	}
	if moduleHasHint(pool, mqIdempotencyHints) {
		return nil, nil
	}
	file, line, ok := firstHintLine(pool, mqConsumeHints)
	if !ok {
		return nil, nil
	}
	return []Finding{finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), file, line)}, nil
}

// broker connection without reconnect/backoff hints
type Mq03 struct{ Base }

func NewMq03() *Mq03 {
	return &Mq03{Base: Base{meta: Meta{ID: "mq-03", Gates: mqGates, Domain: "messaging", DefaultSeverity: SeverityWarn}}}
}

func (c *Mq03) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := mqPool(mod)
	if err != nil {
		return nil, err
	}
	if !moduleHasHint(pool, mqConnectHints) {
		return nil, nil
	}
	if moduleHasHint(pool, mqReconnectHints) {
		return nil, nil
	}
	file, line, ok := firstHintLine(pool, mqConnectHints)
	if !ok {
		return nil, nil
	}
	return []Finding{finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), file, line)}, nil
}

// unbounded goroutine per message in mq consumer path
type Mq04 struct{ Base }

func NewMq04() *Mq04 {
	return &Mq04{Base: Base{meta: Meta{ID: "mq-04", Gates: mqGates, Domain: "messaging", DefaultSeverity: SeverityWarn}}}
}

func (c *Mq04) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := mqPool(mod)
	if err != nil {
		return nil, err
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityWarn)
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		if !fileImportsMQ(f) {
			return true
		}
		forStmt, ok := n.(*ast.ForStmt)
		if !ok {
			return true
		}
		ast.Inspect(forStmt.Body, func(n ast.Node) bool {
			goStmt, ok := n.(*ast.GoStmt)
			if !ok {
				return true
			}
			if _, ok := goStmt.Call.Fun.(*ast.FuncLit); ok {
				findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(goStmt)))
			}
			return true
		})
		return true
	})
	return findings, nil
}

// publish without context deadline in caller
type Mq05 struct{ Base }

func NewMq05() *Mq05 {
	return &Mq05{Base: Base{meta: Meta{ID: "mq-05", Gates: mqGates, Domain: "messaging", DefaultSeverity: SeverityWarn}}}
}

func (c *Mq05) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := mqPool(mod)
	if err != nil {
		return nil, err
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityWarn)
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			return true
		}
		src := string(f.Src)
		if !strings.Contains(src, "WithTimeout") && !strings.Contains(src, "WithDeadline") {
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok || !isMQPublishCall(f, call) {
					return true
				}
				findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(call)))
				return true
			})
		}
		return true
	})
	return findings, nil
}

// kafka reader without consumer group id
type Kafka01 struct{ Base }

func NewKafka01() *Kafka01 {
	return &Kafka01{Base: Base{meta: Meta{ID: "kafka-01", Gates: []string{"kafka-go"}, Domain: "messaging", DefaultSeverity: SeverityFail}}}
}

func (c *Kafka01) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := mqPool(mod)
	if err != nil {
		return nil, err
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityFail)
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		lit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		if !isKafkaReaderConfigLiteral(f, lit) {
			return true
		}
		if readerConfigMissingGroupID(lit) {
			findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(lit)))
		}
		return true
	})
	return findings, nil
}

// kafka commit before process overlaps mq-01
type Kafka02 struct{ Base }

func NewKafka02() *Kafka02 {
	return &Kafka02{Base: Base{meta: Meta{ID: "kafka-02", Gates: []string{"kafka-go"}, Domain: "messaging", DefaultSeverity: SeverityFail}}}
}

func (c *Kafka02) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := mqPool(mod)
	if err != nil {
		return nil, err
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityFail)
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			return true
		}
		if !fileImportsPath(f, kafkaImport) {
			return true
		}
		for _, early := range scanEarlyMQAckInNode(pool, f, fn.Body) {
			if early.id == "mq-01" {
				findings = append(findings, finding(c.ID(), sev, early.file, early.line))
			}
		}
		return true
	})
	return findings, nil
}

// kafka writer message without partition key when ordering may matter
type Kafka03 struct{ Base }

func NewKafka03() *Kafka03 {
	return &Kafka03{Base: Base{meta: Meta{ID: "kafka-03", Gates: []string{"kafka-go"}, Domain: "messaging", DefaultSeverity: SeverityInfo}}}
}

func (c *Kafka03) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := mqPool(mod)
	if err != nil {
		return nil, err
	}
	if !moduleHasHint(pool, []string{"WriteMessages", "kafka.Writer", "kafka.Message"}) {
		return nil, nil
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityInfo)
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		lit, ok := n.(*ast.CompositeLit)
		if !ok || !isKafkaMessageLiteral(f, lit) {
			return true
		}
		if messageLiteralMissingKey(lit) && fileHasAnyHint(f, mqOrderHints) {
			findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(lit)))
		}
		return true
	})
	return findings, nil
}

// shared amqp channel across goroutines
type Rmq01 struct{ Base }

func NewRmq01() *Rmq01 {
	return &Rmq01{Base: Base{meta: Meta{ID: "rmq-01", Gates: []string{"rabbitmq"}, Domain: "messaging", DefaultSeverity: SeverityWarn}}}
}

func (c *Rmq01) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := mqPool(mod)
	if err != nil {
		return nil, err
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityWarn)
	var findings []Finding
	for _, file := range pool.Files {
		if !fileImportsPath(file, amqpImport) {
			continue
		}
		counts := goStmtSharedIdentCounts(file)
		ast.Inspect(file.AST, func(n ast.Node) bool {
			goStmt, ok := n.(*ast.GoStmt)
			if !ok {
				return true
			}
			for _, id := range goStmtIdentArgs(goStmt) {
				if counts[id] >= 2 {
					findings = append(findings, finding(c.ID(), sev, file.RelPath, pool.Line(goStmt)))
					break
				}
			}
			return true
		})
	}
	return findings, nil
}

// queue declare without dead-letter exchange
type Rmq02 struct{ Base }

func NewRmq02() *Rmq02 {
	return &Rmq02{Base: Base{meta: Meta{ID: "rmq-02", Gates: []string{"rabbitmq"}, Domain: "messaging", DefaultSeverity: SeverityWarn}}}
}

func (c *Rmq02) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := mqPool(mod)
	if err != nil {
		return nil, err
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityWarn)
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || !isAmqpQueueDeclare(f, call) {
			return true
		}
		if !queueDeclareHasDLX(call) {
			findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(call)))
		}
		return true
	})
	return findings, nil
}

// amqp dial without reconnect loop heuristic
type Rmq03 struct{ Base }

func NewRmq03() *Rmq03 {
	return &Rmq03{Base: Base{meta: Meta{ID: "rmq-03", Gates: []string{"rabbitmq"}, Domain: "messaging", DefaultSeverity: SeverityWarn}}}
}

func (c *Rmq03) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := mqPool(mod)
	if err != nil {
		return nil, err
	}
	if moduleHasHint(pool, mqReconnectHints) {
		return nil, nil
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityWarn)
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || !isAmqpDial(f, call) {
			return true
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(call)))
		return true
	})
	return findings, nil
}

// core nats subscribe for work that must not be lost
type Nats01 struct{ Base }

func NewNats01() *Nats01 {
	return &Nats01{Base: Base{meta: Meta{ID: "nats-01", Gates: []string{"nats"}, Domain: "messaging", DefaultSeverity: SeverityWarn}}}
}

func (c *Nats01) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := mqPool(mod)
	if err != nil {
		return nil, err
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityWarn)
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		if !fileHasAnyHint(f, natsCriticalHints) {
			return true
		}
		call, ok := n.(*ast.CallExpr)
		if !ok || !isNatsCoreSubscribe(f, call) {
			return true
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(call)))
		return true
	})
	return findings, nil
}

// jetstream consumer without explicit ack handling
type Nats02 struct{ Base }

func NewNats02() *Nats02 {
	return &Nats02{Base: Base{meta: Meta{ID: "nats-02", Gates: []string{"nats"}, Domain: "messaging", DefaultSeverity: SeverityWarn}}}
}

func (c *Nats02) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := mqPool(mod)
	if err != nil {
		return nil, err
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityWarn)
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			return true
		}
		if !funcHasJetStreamConsume(fn) {
			return true
		}
		if funcHasNatsAckHandling(fn) {
			return true
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(fn)))
		return true
	})
	return findings, nil
}

type ackFinding struct {
	id, file string
	line     int
}

func scanEarlyMQAck(pool *astutil.Pool, id, sev string) []Finding {
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			return true
		}
		for _, af := range scanEarlyMQAckInNode(pool, f, fn.Body) {
			if af.id != id {
				continue
			}
			findings = append(findings, finding(id, sev, af.file, af.line))
		}
		return true
	})
	return findings
}

func scanEarlyMQAckInNode(pool *astutil.Pool, f *astutil.File, body ast.Node) []ackFinding {
	var out []ackFinding
	ast.Inspect(body, func(n ast.Node) bool {
		block, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}
		for i, stmt := range block.List {
			if !stmtContainsMQAck(f, stmt) {
				continue
			}
			for j := i + 1; j < len(block.List); j++ {
				if stmtIsMQProcessing(f, block.List[j]) {
					out = append(out, ackFinding{
						id:   "mq-01",
						file: f.RelPath,
						line: pool.Line(stmt),
					})
					break
				}
			}
		}
		return true
	})
	return out
}

func stmtContainsMQAck(f *astutil.File, stmt ast.Stmt) bool {
	found := false
	ast.Inspect(stmt, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || !isMQAckCall(f, call) {
			return true
		}
		found = true
		return false
	})
	return found
}

func stmtIsMQProcessing(f *astutil.File, stmt ast.Stmt) bool {
	found := false
	ast.Inspect(stmt, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			if isMQAckCall(f, node) || isMQNakCall(f, node) {
				return true
			}
			if mqLogCall(f, node) {
				return true
			}
			found = true
			return false
		case *ast.AssignStmt:
			if assignOnlyAckOrErr(node) {
				return true
			}
			found = true
			return false
		case *ast.GoStmt, *ast.DeferStmt:
			found = true
			return false
		}
		return true
	})
	return found
}

func assignOnlyAckOrErr(stmt *ast.AssignStmt) bool {
	if len(stmt.Rhs) != 1 {
		return false
	}
	call, ok := stmt.Rhs[0].(*ast.CallExpr)
	if !ok {
		return false
	}
	return isMQAckCall(nil, call) || isMQNakCall(nil, call)
}

func isMQAckCall(f *astutil.File, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if _, ok := mqAckMethods[sel.Sel.Name]; !ok {
		return false
	}
	if f == nil {
		return true
	}
	if imp, _, ok := selectorImport(f, sel); ok && isMQImport(imp) {
		return true
	}
	return fileImportsMQ(f)
}

func isMQNakCall(f *astutil.File, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Nak" {
		return false
	}
	if f == nil {
		return true
	}
	if imp, _, ok := selectorImport(f, sel); ok && isMQImport(imp) {
		return true
	}
	return fileImportsMQ(f)
}

func mqLogCall(f *astutil.File, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	switch sel.Sel.Name {
	case "Print", "Printf", "Println", "Fatal", "Fatalf", "Error", "Errorf", "Warn", "Warnf", "Info", "Infof", "Debug", "Debugf":
		return true
	default:
		return false
	}
}

func isMQImport(imp string) bool {
	return imp == kafkaImport || imp == amqpImport || imp == natsImport ||
		strings.HasPrefix(imp, kafkaImport+"/") ||
		strings.HasPrefix(imp, amqpImport+"/") ||
		strings.HasPrefix(imp, natsImport+"/")
}

func fileImportsMQ(f *astutil.File) bool {
	for _, imp := range f.Imports {
		if isMQImport(imp) {
			return true
		}
	}
	return false
}

func isMQPublishCall(f *astutil.File, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if _, ok := mqPublishMethods[sel.Sel.Name]; !ok {
		return false
	}
	imp, _, ok := selectorImport(f, sel)
	if !ok {
		return fileImportsMQ(f)
	}
	return isMQImport(imp)
}

func isKafkaReaderConfigLiteral(f *astutil.File, lit *ast.CompositeLit) bool {
	if !fileImportsPath(f, kafkaImport) {
		return false
	}
	switch t := lit.Type.(type) {
	case *ast.SelectorExpr:
		return t.Sel.Name == "ReaderConfig"
	case *ast.Ident:
		return t.Name == "ReaderConfig"
	default:
		return false
	}
}

func readerConfigMissingGroupID(lit *ast.CompositeLit) bool {
	hasGroup := false
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok || key.Name != "GroupID" {
			continue
		}
		hasGroup = true
		if lit, ok := kv.Value.(*ast.BasicLit); ok && lit.Value == `"` {
			return true
		}
		if ident, ok := kv.Value.(*ast.Ident); ok && ident.Name == "" {
			return true
		}
	}
	return !hasGroup
}

func isKafkaMessageLiteral(f *astutil.File, lit *ast.CompositeLit) bool {
	if !fileImportsPath(f, kafkaImport) {
		return false
	}
	switch t := lit.Type.(type) {
	case *ast.SelectorExpr:
		return t.Sel.Name == "Message"
	case *ast.Ident:
		return t.Name == "Message"
	default:
		return false
	}
}

func messageLiteralMissingKey(lit *ast.CompositeLit) bool {
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if ok && key.Name == "Key" {
			return false
		}
	}
	return true
}

func fileHasAnyHint(f *astutil.File, hints []string) bool {
	src := string(f.Src)
	for _, h := range hints {
		if strings.Contains(src, h) {
			return true
		}
	}
	return false
}

func firstHintLine(pool *astutil.Pool, hints []string) (file string, line int, ok bool) {
	for _, f := range pool.Files {
		if astutil.IsTestFile(f.RelPath) {
			continue
		}
		src := string(f.Src)
		for _, hint := range hints {
			if idx := strings.Index(src, hint); idx >= 0 {
				return f.RelPath, strings.Count(src[:idx], "\n") + 1, true
			}
		}
	}
	return "", 0, false
}

func isAmqpQueueDeclare(f *astutil.File, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "QueueDeclare" {
		return false
	}
	imp, _, ok := selectorImport(f, sel)
	return ok && (imp == amqpImport || strings.HasPrefix(imp, amqpImport+"/"))
}

func queueDeclareHasDLX(call *ast.CallExpr) bool {
	for _, arg := range call.Args {
		if tableHasDLX(arg) {
			return true
		}
	}
	return false
}

func tableHasDLX(expr ast.Expr) bool {
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return false
	}
	return tableLiteralHasDLX(lit)
}

func tableLiteralHasDLX(lit *ast.CompositeLit) bool {
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		if keyLit, ok := kv.Key.(*ast.BasicLit); ok && keyLit.Kind == token.STRING {
			val := strings.Trim(keyLit.Value, `"`)
			if strings.Contains(val, "dead-letter") {
				return true
			}
		}
	}
	return false
}

func isAmqpDial(f *astutil.File, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Dial" {
		return false
	}
	imp, _, ok := selectorImport(f, sel)
	return ok && (imp == amqpImport || strings.HasPrefix(imp, amqpImport+"/"))
}

func goStmtSharedIdentCounts(f *astutil.File) map[string]int {
	counts := make(map[string]int)
	ast.Inspect(f.AST, func(n ast.Node) bool {
		goStmt, ok := n.(*ast.GoStmt)
		if !ok {
			return true
		}
		seen := make(map[string]struct{})
		for _, id := range goStmtIdentArgs(goStmt) {
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			counts[id]++
		}
		return true
	})
	return counts
}

func goStmtIdentArgs(goStmt *ast.GoStmt) []string {
	var out []string
	ast.Inspect(goStmt, func(n ast.Node) bool {
		ident, ok := n.(*ast.Ident)
		if !ok || ident.Name == "_" {
			return true
		}
		out = append(out, ident.Name)
		return true
	})
	return out
}

func isNatsCoreSubscribe(f *astutil.File, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Subscribe" {
		return false
	}
	if !fileImportsPath(f, natsImport) {
		return false
	}
	if fileHasAnyHint(f, jetStreamHints) {
		return false
	}
	return true
}

func funcHasJetStreamConsume(fn *ast.FuncDecl) bool {
	if fn.Body == nil {
		return false
	}
	found := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		for _, hint := range jetStreamHints {
			if sel.Sel.Name == hint || strings.Contains(sel.Sel.Name, hint) {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

func funcHasNatsAckHandling(fn *ast.FuncDecl) bool {
	if fn.Body == nil {
		return false
	}
	found := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.Ident:
			for _, hint := range natsAckHints {
				if v.Name == hint || strings.Contains(v.Name, hint) {
					found = true
					return false
				}
			}
		case *ast.SelectorExpr:
			for _, hint := range natsAckHints {
				if v.Sel.Name == hint || strings.Contains(v.Sel.Name, hint) {
					found = true
					return false
				}
			}
		}
		return true
	})
	return found
}
