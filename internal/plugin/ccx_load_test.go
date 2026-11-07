package plugin

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestCCXPluginLoads 验证 ccx 插件能被 Magpie 的宿主加载。
func TestCCXPluginLoads(t *testing.T) {
	root := t.TempDir()
	authPath := filepath.Join(root, "auth.json")
	os.WriteFile(authPath, []byte("{}"), 0o600)

	h := &host{
		in:       discardCloser{},
		calls:    map[int64]*call{},
		dead:     make(chan struct{}),
		slots:    make(chan struct{}, callLimit()),
		streams:  make(chan struct{}, streamLimit()),
		out:      make(chan writeReq, callLimit()),
		ctrl:     map[int64]*ctrl{},
		ctrlWake: make(chan struct{}, 1),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 模拟 init 调用
	params := map[string]any{
		"authPath":      authPath,
		"directory":     root,
		"modelsDevPath": "",
		"plugins": []map[string]any{{
			"spec":   "ccx",
			"target": "/Users/liu/magpie/plugins/ccx",
		}},
		"loadMillis": 5000,
		"listMillis": 5000,
	}
	var res struct {
		Plugins []Loaded `json:"plugins"`
	}
	if err := h.call(ctx, "init", params, &res); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if len(res.Plugins) == 0 {
		t.Fatal("no plugins loaded")
	}
	if res.Plugins[0].Error != "" {
		t.Fatalf("plugin error: %s", res.Plugins[0].Error)
	}
	b, _ := json.Marshal(res.Plugins[0])
	t.Logf("loaded: %s", b)
}
