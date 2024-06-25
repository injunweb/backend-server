package security

import (
	"sync"

	"github.com/injunweb/backend-server/internal/global/utils"
)

type SecurityContext struct {
	ID string
}

type ContextHolder struct {
	mu       sync.Mutex
	contexts map[uint64]*SecurityContext
}

var holder *ContextHolder

func init() {
	holder = NewContextHolder()
}

func NewContextHolder() *ContextHolder {
	return &ContextHolder{
		contexts: make(map[uint64]*SecurityContext),
	}
}

func SetContext(ctx *SecurityContext) {
	holder.SetContext(ctx)
}

func GetContext() *SecurityContext {
	return holder.GetContext()
}

func RemoveContext() {
	holder.RemoveContext()
}

func (ch *ContextHolder) SetContext(ctx *SecurityContext) {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	ch.contexts[utils.GetGoroutineID()] = ctx
}

func (ch *ContextHolder) GetContext() *SecurityContext {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	return ch.contexts[utils.GetGoroutineID()]
}

func (ch *ContextHolder) RemoveContext() {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	delete(ch.contexts, utils.GetGoroutineID())
}
