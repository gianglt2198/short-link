package middlewares

import (
	"github.com/gianglt1/short-link/internal/common"
	"github.com/gianglt1/short-link/internal/config"
	"github.com/gianglt1/short-link/internal/infra/logging"
	"github.com/gianglt1/short-link/internal/utils"
	fiber "github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func RequestIDMiddleware(c *fiber.Ctx) error {
	ctx, requestID := utils.ApplyRequestIDWithContext(c.UserContext())
	c.Locals(common.KEY_REQUEST_ID, requestID)
	c.SetUserContext(ctx)
	return c.Next()
}

func LoggerMiddleware(logger *logging.Logger, cfg *config.Config) fiber.Handler {
	skip := make(map[string]struct{}, len(cfg.Logging.SkipPaths))
	for _, p := range cfg.Logging.SkipPaths {
		skip[p] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		if _, ok := skip[c.Path()]; ok {
			return c.Next()
		}

		logFields := []zapcore.Field{
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.String("request_id", c.Locals(common.KEY_REQUEST_ID).(string)),
		}
		if err := c.Next(); err != nil {
			logger.GetLogger().With(logFields...).Error("Request failed", zap.Error(err))
			return err
		}

		logger.GetLogger().With(logFields...).Info("Request succeeded")
		return nil
	}
}
