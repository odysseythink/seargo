// Package answerer provides instant-answer backends for direct query responses.
//
// Answerers are keyword-triggered: when the first word of a query matches
// an answerer's keywords, AnswerStorage.Ask dispatches to the answerer.
//
// Built-in answerers are registered via init() in internal/answerer/builtin/.
package answerer
