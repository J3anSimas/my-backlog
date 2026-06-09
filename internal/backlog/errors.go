package backlog

import "my-backlog/internal/apperrors"

// InputKind, InputError e InfraError são aliases dos tipos canônicos em apperrors.
// Código existente que referencia backlog.InputError continua funcionando.
type InputKind = apperrors.InputKind
type InputError = apperrors.InputError
type InfraError = apperrors.InfraError

const (
	KindInvalid  = apperrors.KindInvalid
	KindNotFound = apperrors.KindNotFound
)
