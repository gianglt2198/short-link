package middlewares

import (
	"github.com/gianglt1/short-link/internal/common"
	"github.com/gianglt1/short-link/internal/infra/logging"
	"github.com/gianglt1/short-link/internal/utils"
	fiber "github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const ()

func RequestIDMiddleware(c *fiber.Ctx) error {
	ctx, requestID := utils.ApplyRequestIDWithContext(c.UserContext())
	c.Locals(common.KEY_REQUEST_ID, requestID)
	c.SetUserContext(ctx)
	return c.Next()
}

func LoggerMiddleware(logger *logging.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
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
