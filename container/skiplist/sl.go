package skiplist

import (
	"math/rand"
	"sync"
	"time"
)

const (
	DefaultMaxLevel    = 16  // 最大层数
	DefaultProbability = 0.5 // 晋升概率 1/2
)

// Node 跳表节点
type Node[T any] struct {
	value T          // 节点值
	level []*Node[T] // 每一层的前向指针
}

// SkipList 泛型跳表
type SkipList[T any] struct {
	head        *Node[T]          // 头节点（哨兵）
	better      func(a, b T) bool // 比较函数：返回 true 表示 a 排在 b 前面
	maxLevel    int               // 最大层数
	probability float64           // 晋升概率
	level       int               // 当前最大层数
	length      int               // 元素数量
	mu          *sync.RWMutex     // 读写锁，保证并发安全
	rng         *rand.Rand        // 随机数生成器
}

// ═══════════════════════════════════════════════════════
// 构造函数
// ═══════════════════════════════════════════════════════

// New 创建新的跳表
// better: 比较函数，返回 true 表示 a 排在 b 前面（即 a 更小/优先级更高）
func New[T any](better func(a, b T) bool, withRWLock bool) *SkipList[T] {
	return NewWithConfig(better, DefaultMaxLevel, DefaultProbability, withRWLock)
}

// NewWithConfig 使用自定义配置创建跳表
func NewWithConfig[T any](better func(a, b T) bool, maxLevel int, probability float64, withRWLock bool) *SkipList[T] {
	var rw_lock *sync.RWMutex
	if withRWLock {
		rw_lock = new(sync.RWMutex)
	}

	sl := &SkipList[T]{
		better:      better,
		maxLevel:    maxLevel,
		probability: probability,
		level:       1,
		length:      0,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
		mu:          rw_lock,
	}

	// 创建头节点，头节点的值不使用，但需要初始化所有层的前向指针
	sl.head = &Node[T]{
		level: make([]*Node[T], maxLevel),
	}

	return sl
}

// ═══════════════════════════════════════════════════════
// 辅助方法
// ═══════════════════════════════════════════════════════

// randomLevel 随机生成节点层数
// 使用概率 1/2，即每层晋升概率为 0.5
func (sl *SkipList[T]) randomLevel() int {
	level := 1
	for sl.rng.Float64() < sl.probability && level < sl.maxLevel {
		level++
	}
	return level
}

// createNode 创建新节点
func (sl *SkipList[T]) createNode(value T, level int) *Node[T] {
	return &Node[T]{
		value: value,
		level: make([]*Node[T], level),
	}
}

// compare 比较两个值
// 返回值：
//
//	-1: a 排在 b 前面（a < b）
//	 0: a 等于 b
//	 1: a 排在 b 后面（a > b）
func (sl *SkipList[T]) compare(a, b T) int {
	if sl.better(a, b) {
		return -1
	}
	if sl.better(b, a) {
		return 1
	}
	return 0
}

// findPredecessors 找到每一层的前驱节点
// 用于插入和删除操作
func (sl *SkipList[T]) findPredecessors(value T) []*Node[T] {
	predecessors := make([]*Node[T], sl.maxLevel)
	current := sl.head

	// 从最高层开始向下搜索
	for i := sl.level - 1; i >= 0; i-- {
		// 在当前层向右移动，直到找到合适的位置
		for current.level[i] != nil && sl.better(current.level[i].value, value) {
			current = current.level[i]
		}
		predecessors[i] = current
	}

	return predecessors
}

// ═══════════════════════════════════════════════════════
// 核心操作
// ═══════════════════════════════════════════════════════

// Insert 插入元素（并发安全）
func (sl *SkipList[T]) Insert(value T) {
	if sl.mu != nil {
		sl.mu.Lock()
		defer sl.mu.Unlock()
	}

	// 找到每一层的前驱节点
	predecessors := sl.findPredecessors(value)

	// 检查是否已存在（去重）
	firstLevelPredecessor := predecessors[0]
	if firstLevelPredecessor.level[0] != nil {
		if sl.compare(firstLevelPredecessor.level[0].value, value) == 0 {
			return // 已存在，不重复插入
		}
	}

	// 随机生成新节点的层数
	newLevel := sl.randomLevel()

	// 如果新节点的层数大于当前跳表的层数，需要更新跳表层数
	if newLevel > sl.level {
		for i := sl.level; i < newLevel; i++ {
			predecessors[i] = sl.head
		}
		sl.level = newLevel
	}

	// 创建新节点
	newNode := sl.createNode(value, newLevel)

	// 在每一层插入新节点
	for i := range newLevel {
		newNode.level[i] = predecessors[i].level[i]
		predecessors[i].level[i] = newNode
	}

	sl.length++
}

// Search 查找元素（并发安全）
func (sl *SkipList[T]) Search(value T) *Node[T] {
	if sl.mu != nil {
		sl.mu.RLock()
		defer sl.mu.RUnlock()
	}

	current := sl.head

	// 从最高层开始搜索
	for i := sl.level - 1; i >= 0; i-- {
		// 在当前层向右移动
		for current.level[i] != nil && sl.better(current.level[i].value, value) {
			current = current.level[i]
		}
	}

	// 移动到第 0 层的下一个节点
	current = current.level[0]

	// 检查是否找到
	if current != nil && sl.compare(current.value, value) == 0 {
		return current
	}

	return nil
}

// Delete 删除元素（并发安全）
func (sl *SkipList[T]) Delete(value T) bool {
	if sl.mu != nil {
		sl.mu.Lock()
		defer sl.mu.Unlock()
	}

	// 找到每一层的前驱节点
	predecessors := sl.findPredecessors(value)

	// 检查节点是否存在
	target := predecessors[0].level[0]
	if target == nil || sl.compare(target.value, value) != 0 {
		return false
	}

	// 在每一层删除节点
	for i := 0; i < len(target.level); i++ {
		if predecessors[i].level[i] == target {
			predecessors[i].level[i] = target.level[i]
		}
	}

	// 更新跳表的层数（如果最高层没有节点了）
	for sl.level > 1 && sl.head.level[sl.level-1] == nil {
		sl.level--
	}

	sl.length--
	return true
}

// ═══════════════════════════════════════════════════════
// 查询操作
// ═══════════════════════════════════════════════════════

// RangeQuery 范围查询（并发安全）
// 返回 [min, max] 范围内的所有元素
func (sl *SkipList[T]) RangeQuery(min, max T) []T {
	if sl.mu != nil {
		sl.mu.RLock()
		defer sl.mu.RUnlock()
	}

	result := make([]T, 0)

	// 找到 min 的位置（或第一个不小于 min 的节点）
	current := sl.head
	for i := sl.level - 1; i >= 0; i-- {
		for current.level[i] != nil && sl.better(current.level[i].value, min) {
			current = current.level[i]
		}
	}

	// 移动到第一个可能符合条件的节点
	current = current.level[0]

	// 遍历收集范围内的元素
	for current != nil {
		// 如果当前值已经超过 max，停止遍历
		if sl.better(max, current.value) {
			break
		}
		result = append(result, current.value)
		current = current.level[0]
	}

	return result
}

// GetMin 获取最小元素（并发安全）
func (sl *SkipList[T]) GetMin() T {
	if sl.mu != nil {
		sl.mu.RLock()
		defer sl.mu.RUnlock()
	}

	var zero T
	if sl.length == 0 {
		return zero
	}

	return sl.head.level[0].value
}

// GetMax 获取最大元素（并发安全）
func (sl *SkipList[T]) GetMax() T {
	if sl.mu != nil {
		sl.mu.RLock()
		defer sl.mu.RUnlock()
	}

	var zero T
	if sl.length == 0 {
		return zero
	}

	// 从最高层开始向下找最后一个节点
	current := sl.head
	for i := sl.level - 1; i >= 0; i-- {
		for current.level[i] != nil {
			current = current.level[i]
		}
	}

	return current.value
}

// ═══════════════════════════════════════════════════════
// 辅助操作
// ═══════════════════════════════════════════════════════

// Len 获取元素数量（并发安全）
func (sl *SkipList[T]) Len() int {
	if sl.mu != nil {
		sl.mu.RLock()
		defer sl.mu.RUnlock()
	}

	return sl.length
}

// Contains 检查元素是否存在
func (sl *SkipList[T]) Contains(value T) bool {
	return sl.Search(value) != nil
}

// GetAll 获取所有元素（按序）
func (sl *SkipList[T]) GetAll() []T {
	if sl.mu != nil {
		sl.mu.RLock()
		defer sl.mu.RUnlock()
	}

	result := make([]T, 0, sl.length)
	current := sl.head.level[0]

	for current != nil {
		result = append(result, current.value)
		current = current.level[0]
	}

	return result
}

// Clear 清空跳表
func (sl *SkipList[T]) Clear() {
	if sl.mu != nil {
		sl.mu.Lock()
		defer sl.mu.Unlock()
	}

	sl.head = &Node[T]{
		level: make([]*Node[T], sl.maxLevel),
	}
	sl.level = 1
	sl.length = 0
}
