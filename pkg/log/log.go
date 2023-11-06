// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"context"

	"go.uber.org/zap"
)

type loggerKeyType string

var loggerKey loggerKeyType = "logger"

// NewContext returns a new context with a *zap.SugaredLogger and panics if the logger is nil.
func NewContext(parentCtx context.Context, logger *zap.SugaredLogger) context.Context {
	if logger == nil {
		panic("Nil logger")
	}
	return context.WithValue(parentCtx, loggerKey, logger)
}

// FromContext retrieves a *zap.SugaredLogger from a context and panics if there isn't one.
func FromContext(ctx context.Context) *zap.SugaredLogger {
	val := ctx.Value(loggerKey)
	logger, ok := val.(*zap.SugaredLogger)
	if !ok {
		panic("No logger in context")
	}
	return logger
}
