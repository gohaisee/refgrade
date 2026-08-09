package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gohaisee/refgrade/fixtures/bad-mass-assign/internal/model"
)

func Update(w http.ResponseWriter, r *http.Request) {
	_ = json.Unmarshal([]byte("{}"), &model.User{})
}
