package logger

import "go.uber.org/zap"

var Logger *zap.SugaredLogger

func init() {
	sugar := zap.NewExample().Sugar()
	defer sugar.Sync()
	Logger = sugar
}