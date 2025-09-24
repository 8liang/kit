package event

import (
	"sync"
	"sync/atomic"
)

// NewListener 创建新的监听器实例（默认启用线程安全）
func NewListener() *Listener {
	return NewListenerWithConfig(Config{ThreadSafe: true})
}

// NewListenerUnsafe 创建非线程安全的监听器实例（性能更好）
func NewListenerUnsafe() *Listener {
	return NewListenerWithConfig(Config{ThreadSafe: false})
}

// NewListenerWithConfig 使用配置创建监听器实例
func NewListenerWithConfig(config Config) *Listener {
	return &Listener{
		handlers:   make(map[Type][]HandlerInfo),
		nextID:     0,
		threadSafe: config.ThreadSafe,
	}
}

// HandlerInfo 存储处理函数和其唯一标识符
type HandlerInfo struct {
	ID      uint64
	Handler func(eventData *Event)
}

// Listener 事件监听器，支持多个监听器和删除功能
type Listener struct {
	handlers   map[Type][]HandlerInfo
	nextID     uint64
	mu         sync.RWMutex
	threadSafe bool
}

// On 添加事件监听器，返回唯一ID用于删除
func (l *Listener) On(eventType Type, handler func(eventData *Event)) uint64 {
	if l.threadSafe {
		l.mu.Lock()
		defer l.mu.Unlock()
	}

	id := atomic.AddUint64(&l.nextID, 1)
	handlerInfo := HandlerInfo{
		ID:      id,
		Handler: handler,
	}

	l.handlers[eventType] = append(l.handlers[eventType], handlerInfo)
	return id
}

// Off 删除特定的监听器
func (l *Listener) Off(eventType Type, handlerID uint64) bool {
	if l.threadSafe {
		l.mu.Lock()
		defer l.mu.Unlock()
	}

	handlers, exists := l.handlers[eventType]
	if !exists {
		return false
	}

	for i, handler := range handlers {
		if handler.ID == handlerID {
			// 删除指定的处理器
			l.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			// 如果没有处理器了，删除整个键
			if len(l.handlers[eventType]) == 0 {
				delete(l.handlers, eventType)
			}
			return true
		}
	}

	return false
}

// OffAll 删除某个事件类型的所有监听器
func (l *Listener) OffAll(eventType Type) {
	if l.threadSafe {
		l.mu.Lock()
		defer l.mu.Unlock()
	}

	delete(l.handlers, eventType)
}

// Emit 触发事件，按注册顺序执行所有监听器
func (l *Listener) Emit(event *Event) {
	var handlersCopy []HandlerInfo

	if l.threadSafe {
		l.mu.RLock()
		handlers, exists := l.handlers[event.Type()]
		if !exists {
			l.mu.RUnlock()
			return
		}
		// 复制处理器切片以避免在执行过程中被修改
		handlersCopy = make([]HandlerInfo, len(handlers))
		copy(handlersCopy, handlers)
		l.mu.RUnlock()
	} else {
		handlers, exists := l.handlers[event.Type()]
		if !exists {
			return
		}
		// 非线程安全模式下直接使用原切片
		handlersCopy = handlers
	}

	// 按注册顺序执行所有处理器
	for _, handler := range handlersCopy {
		handler.Handler(event)
	}
}

// Count 返回指定事件类型的监听器数量
func (l *Listener) Count(eventType Type) int {
	if l.threadSafe {
		l.mu.RLock()
		defer l.mu.RUnlock()
	}

	return len(l.handlers[eventType])
}

// HasListeners 检查是否有监听器注册到指定事件类型
func (l *Listener) HasListeners(eventType Type) bool {
	if l.threadSafe {
		l.mu.RLock()
		defer l.mu.RUnlock()
	}

	handlers, exists := l.handlers[eventType]
	return exists && len(handlers) > 0
}

func Cast[T any](e *Event) (T, bool) {
	if data, ok := e.Payload().(T); ok {
		return data, true
	}
	var zero T
	return zero, false
}

func OnTyped[T any](l *Listener, tp Type, call func(d T)) {
	l.On(tp, func(eventIns *Event) {
		if payload, ok := Cast[T](eventIns); ok {
			call(payload)
		}
	})
}
