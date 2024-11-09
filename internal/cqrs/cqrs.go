package cqrs

type CommandHandler[TIn any] interface {
	Execute(cmd TIn) error
}

type CommandHandlerWithResponse[TIn any, TOut any] interface {
	Execute(cmd TIn) (TOut, error)
}
