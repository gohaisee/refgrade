package detect_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/gohaisee/refgrade/internal/detect"
	"github.com/gohaisee/refgrade/internal/project"
)

func TestDetect_findsMongoDriver(t *testing.T) {
	t.Parallel()
	root := filepath.Join("..", "..", "testdata", "fixtures", "bad-mongo-per-request")
	mod, err := project.Load(context.Background(), root, project.LoadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	stacks, err := detect.Detect(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range stacks {
		if s.Name == "mongo-driver" {
			return
		}
	}
	t.Fatalf("stacks = %+v", stacks)
}
