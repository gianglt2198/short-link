package helpers

import "go.uber.org/fx"

var Module = fx.Provide(
	fx.Annotate(NewSnowflakeCodeGenerator, fx.As(new(CodeGenerator))),
)
