package main

import (
	"database/sql"
	"expvar"
	"fmt"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

func expVarHandler(metricToSkip map[string]any) gin.HandlerFunc {
	return func(c *gin.Context) {
		w := c.Writer
		c.Header("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte("{\n"))
		skip := true
		expvar.Do(func(kv expvar.KeyValue) {
			if !skip {
				_, _ = w.Write([]byte(",\n"))
			}
			if _, ok := metricToSkip[kv.Key]; !ok {
				skip = false
				fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
			} else {
				skip = true
			}
		})
		_, _ = w.Write([]byte("\n}\n"))
		c.AbortWithStatus(200)
	}
}

func customMetric(db *sql.DB) {
	expvar.NewString("version").Set(version)
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))
	expvar.Publish("db", expvar.Func(func() any {
		return db.Stats()
	}))
	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().UTC().Unix()
	}))
}
