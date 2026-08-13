package bst

import "io"

type CMPFN[T any] = func(a, b T) int

// BSTNodeInterface[T] 是「BST 节点」行为的最小契约,供同包内 SelfBalancingBST 的实现使用。
// 实现者(treap / splay / avl / ...)自行决定具体结构体,只要满足下列不变量:
//
//  1. 根节点的 GetParent() 必须返回 nil (约定:parent == nil 即根,无外部哨兵)
//  2. 任意非根节点 N 必须满足 N.GetParent() != nil 且
//     (N.GetParent().GetLeft() == N || N.GetParent().GetRight() == N)
//  3. 调用 SetLeft / SetRight 后,实现者必须**自行**对新孩子调用 SetParent(this);
//     反之亦然。否则 rotate / Get / Delete 会读到脏数据(rotater 仅同步旋转局部的指针)
//  4. SetSize 之后,本节点 size 应当与左右孩子的 size 一致(由 UpdateSize 强制保证);
//     沿父链的 size 刷新由调用方负责
//  5. 节点必须持有一个由所属 SelfBalancingBST 在创建时透传过来的比较函数;
//     节点的 Compare(a, b) 必须与该 tree 的 IsLessThan()(a, b) 返回完全一致
//  6. 节点持有的比较器在节点生命周期内不可替换;tree 若需换比较器,应创建新 tree
//  7. 若 impl 嵌入 *Node[T],则 **不可** 在具体类型上重写 BSTNodeInterface 的 method
//     (不允许定义 treapNode.GetVal 等同名 method,Go 会用具体类型版本,
//     绕过 *Node[T] 的逻辑,容易破坏 size/cmp 不变量)
//  8. 嵌入指针 *Node[T] 在具体节点生命周期内不可替换
//     (类似 Compare 不可替换,避免 history-confusion)
//
// 线程安全:BSTNodeInterface 实例不保证自身线程安全;并发由所属的 SelfBalancingBST 实现决定。
//
// 推荐实现方式(嵌入基类 Node[T]):
//
//	type myNode[T any] struct {
//	    *Node[T]      // 嵌入,自动获得全部 10+1 个 BSTNodeInterface method
//	    // impl 自己的字段(priority / color / ...)
//	}
//
// 嵌入后 *myNode[T] 自动满足 BSTNodeInterface[T],无需手写任何转发 method。
// 禁止在 impl 中重写 BSTNodeInterface 的 method(否则会掩盖嵌入基类的版本)。
type BSTNodeInterface[T any] interface {
	GetVal() T
	GetLeft() BSTNodeInterface[T]
	GetRight() BSTNodeInterface[T]
	GetParent() BSTNodeInterface[T]

	SetLeft(BSTNodeInterface[T])
	SetRight(BSTNodeInterface[T])
	SetParent(BSTNodeInterface[T])

	GetSize() int
	SetSize(int)

	// CompareFn 用节点自身持有的比较器对 a, b 做三态比较:
	//   -1 表示 a < b;0 表示 a == b;1 表示 a > b。
	// 不可变性:比较器在节点生命周期内不得替换。
	CompareFn() func(a, b T) int
}

var _ BSTNodeInterface[int] = new(Node[int])

// Node[T] 是 BSTNodeInterface[T] 接口的默认实现,供具体 tree 节点(treap / splay / avl / ...)嵌入使用。
//
// 设计目标:把 10+1 个 BSTNodeInterface method 的样板代码集中到一处,具体 impl 只需要
//
//	type treapNode[T any] struct {
//	    *Node[T]   // 嵌入指针,所有 BSTNodeInterface method 会被 method promotion 提升到 *treapNode[T]
//	    priority int
//	}
//
// 即可自动满足 BSTNodeInterface[T] 接口,无需手写任何转发 method。
//
// 字段语义:
//   - val: 节点存储的元素值
//   - left / right: 左右孩子(interface 类型,允许树自身递归)
//   - parent: 父节点;根节点此字段为 nil
//   - size: 子树大小(含自身);约定空子树 size = 0
//   - cmp: 由所属 SelfBalancingBST 在创建时透传过来的三态比较函数(返回 -1 / 0 / 1)
//
// 嵌入方式必须是 *Node[T](指针嵌入),不能用 Node[T](值嵌入),
// 否则 method receiver 类型不对,无法被 BSTNodeInterface 接口接受。
//
// 字段访问:由于 *Node[T] 是嵌入字段,*treapNode[T] 可直接写 tn.val / tn.size / tn.cmp
// 访问基类字段(method promotion 同样适用于字段)。
type Node[T any] struct {
	val    T
	left   BSTNodeInterface[T]
	right  BSTNodeInterface[T]
	parent BSTNodeInterface[T]
	size   int
	cmp    func(a, b T) int
}

// NewNode 创建并初始化一个 *Node[T]。
//   - val: 节点值
//   - cmp: 三态比较函数(返回 -1 / 0 / 1)
//
// 创建后 size 默认为 1(单节点子树);left / right / parent 默认为 nil。
// 具体 impl(treap / splay / avl)的构造工厂应当:
//  1. 调用 NewNode 拿到 *Node[T]
//  2. 用其初始化 *具体Node[T]
//  3. 写入具体 impl 自己的字段(priority / color / ...)
func NewNode[T any](val T, cmp func(a, b T) int) *Node[T] {
	return &Node[T]{
		val:  val,
		size: 1,
		cmp:  cmp,
	}
}

// 下面 11 个 method 共同让 *Node[T] 满足 BSTNodeInterface[T]。
// 具体 impl 嵌入 *Node[T] 后,通过 method promotion 自动继承。

func (n *Node[T]) GetVal() T                      { return n.val }
func (n *Node[T]) GetLeft() BSTNodeInterface[T]   { return n.left }
func (n *Node[T]) GetRight() BSTNodeInterface[T]  { return n.right }
func (n *Node[T]) GetParent() BSTNodeInterface[T] { return n.parent }

func (n *Node[T]) SetLeft(b BSTNodeInterface[T])  { n.left = b }
func (n *Node[T]) SetRight(b BSTNodeInterface[T]) { n.right = b }
func (n *Node[T]) SetParent(b BSTNodeInterface[T]) {
	n.parent = b
}

func (n *Node[T]) GetSize() int  { return n.size }
func (n *Node[T]) SetSize(s int) { n.size = s }

func (n *Node[T]) CompareFn() func(a, b T) int { return n.cmp }


func (n *Node[T]) leftNode() *Node[T] {
	if l := n.GetLeft(); l != nil {
		return l.(*Node[T])
	}
	return nil
}

func (n *Node[T]) rightNode() *Node[T] {
	if r := n.GetRight(); r != nil {
		return r.(*Node[T])
	}
	return nil
}

func (n *Node[T]) parentNode() *Node[T] {
	if pp := n.GetParent(); pp != nil {
		return pp.(*Node[T])
	}
	return nil
}

// BST 有序约定（所有实现/工具必须一致）：
//   - 对任意节点 n，左子树任意节点 l 满足 cmp(n, l) >= 0（允许相等）；
//   - 右子树任意节点 r 满足 cmp(n, r) < 0（严格小于）。
//   - 相等元素允许共存，且插入时走左子树。
// 实现者（Insert/Delete/Update/Get 等）必须遵守此约定，否则破坏有序不变量。
type SelfBalancingBST[T any] interface {

	// Insert
	// 功能：插入结点。当树内已经存在和 item 相等的元素时，不覆盖旧元素，
	// 而是把 item 作为一个新节点添加进树中（允许重复元素共存）。
	// 平衡树内部自动执行对应的平衡维护(旋转 / splay / 重建 / treap 堆调整等)
	// 返回值：恒为 true（始终新增一个节点）。
	//
	// 重复元素策略：相等元素(cmp == 0)允许共存，不覆盖。
	//
	// 示例场景：
	// 1. 记录一批 key 相同但 value 不同的日志/事件，全部保留不覆盖；
	// 2. 多重集(multiset)语义的有序容器。
	Insert(item T)

	// Update
	// 功能：定位与 item「相等」的节点（按 cmp 判定），调用 callback 修改其内容后，
	// 自动检测排序键是否变化并重新调整位置，使树保持有序。
	//  - item 用样板（同 cmp 语义）定位；典型 T 为指针，cmp 先比较指针身份。
	//  - callback 修改节点内容（如改变参与 cmp 的字段）。
	//  - 若修改后排序键变化（不再满足 BST 有序不变量），实现需将节点移到正确位置。
	//  - 若树中不存在与 item 相等的节点，本方法无副作用（不回调、不新增）。
	//
	//
	// 示例场景：
	// 1. 排行榜：玩家得分更新后，重新按新得分排序；
	// 2. 调度器：任务的触发时间更新后，重新入队。
	Update(item T, callback func(item T))

	// Get
	// 功能：传入样板item，查找树中和它相等的结点；返回存储的原始元素以及是否命中
	// 等值依据:IsLessThan()(item, hit) == 0
	// 特殊行为(Splay‑Tree)：查询成功之后自动将访问过的结点伸展至树根
	//
	// 示例场景：
	// 1. 根据id样板查找用户完整结构体；
	// 2. 查询某个定时任务是否还存放在任务集合；
	// 3. 缓存结构读取一条记录，splay‑tree 借此优化热点访问。
	Get(item T) (T, bool)

	// Delete
	// 功能：删除和入参item相等的节点；删除之后自动维护树平衡
	// 返回值：true = 元素成功移除；false = 找不到目标
	//
	// 示例场景：
	// 1. 定时任务已经执行完毕，从任务优先级树删除；
	// 2. 用户注销，从有序用户集合移除对应条目；
	// 3. LRU‑Tree 删除已经过期的数据结点。
	Delete(item T) bool

	// Size
	// 获取当前树里面存储的元素总个数
	//
	// 示例场景：
	// 1. 判断有序集合一共有多少条在线用户；
	// 2. 前置校验 Select(rank) 的下标是否越界；
	// 3. 监控优先级队列任务总量。
	Size() int

	// IsEmpty
	// 判断树当中是否不存在任何元素
	//
	// 示例场景：
	// 1. 循环消费定时任务，!IsEmpty() 就取出最小任务；
	// 2. 程序启动校验有序缓存是否为空。
	IsEmpty() bool

	// Clear
	// 清空全部节点，重置为空树，释放引用；比较函数保持不变
	//
	// 示例场景：
	// 1. 配置重载，清空旧版本有序配置集合；
	// 2. 连接重置，清空会话树。
	Clear()

	// InOrderTraverse
	// 中序遍历，严格依照 IsLessThan 定义的升序顺序遍历全部元素
	// 回调 fn 接收每一个遍历到的元素；回调内禁止修改树结构
	//
	// 示例场景：
	// 1. 将全部用户按年龄升序打印；
	// 2. 持久化，顺序导出整棵有序集合的数据；
	// 3. 迭代所有待执行任务依次处理。
	InOrderTraverse(fn func(T))

	// PreorderTraverse
	// 前序遍历：先访问当前节点，再遍历左子树、右子树。
	// 回调 fn 接收每一个访问到的元素；回调内禁止修改树结构。
	//
	// 示例场景：
	// 1. 深拷贝树的节点结构（先父后子）；
	// 2. 按前缀顺序导出/序列化结构数据。
	PreorderTraverse(fn func(T))

	// PostorderTraverse
	// 后序遍历：先遍历左子树、右子树，最后访问当前节点。
	// 回调 fn 接收每一个访问到的元素；回调内禁止修改树结构。
	//
	// 示例场景：
	// 1. 自底向上计算/回收（子先于父）；
	// 2. 删除整棵树时先处理子树再处理父节点。
	PostorderTraverse(fn func(T))

	// Min
	// 获取排序序列当中排在首位、IsLessThan 判断为最小的元素
	// 返回 (zero‑value, false) 当树为空
	//
	// 示例场景：
	// 1. 优先队列取出最近需要触发的定时任务；
	// 2. 获取年龄最小的用户。
	Min() (T, bool)

	// Max
	// 获取排序序列当中最大、排在序列末尾的元素
	//
	// 示例场景：
	// 1. 获取最晚到期的超时会话；
	// 2. 获取排行榜分数最高的条目。
	Max() (T, bool)

	// Predecessor
	// 查询：严格小于 item 的所有元素里面最大那一个
	//  - 树为空 / 不存在这样的前驱(例如 item 已是最小):返回 (zero, false)
	//  - 若 item 恰好存在于树中,前驱严格小于 item,不会返回 item 自身
	//  - 小于判定依据:IsLessThan()(a, item) == -1
	//
	// 特殊行为(Splay‑Tree):本方法内部需要定位 item 所在位置,实现可能
	// 在定位过程中把访问过的节点伸展至根(副作用)。treap / AVL 等无此副作用。
	//
	// 示例场景：
	// 1. 在时间轴上找到当前任务的上一个定时任务；
	// 2. 分数排行榜，查找一名玩家的上一名玩家；
	// 3. 区间分页，获取上一页边界。
	Predecessor(item T) (T, bool)

	// Successor
	// 查询：严格大于 item 的所有元素里面最小那一个
	//  - 树为空 / 不存在这样的后继(例如 item 已是最大):返回 (zero, false)
	//  - 若 item 恰好存在于树中,后继严格大于 item,不会返回 item 自身
	//  - 大于判定依据:IsLessThan()(item, a) == -1
	//
	// 特殊行为(Splay‑Tree):本方法内部需要定位 item 所在位置,实现可能
	// 在定位过程中把访问过的节点伸展至根(副作用)。treap / AVL 等无此副作用。
	//
	// 示例场景：
	// 1. 找到大于当前时间的最近一次定时任务；
	// 2. 分页查找下一个边界元素；
	// 3. 区间缺口检测，检查相邻两个数值之间空隙。
	Successor(item T) (T, bool)

	// RangeVisit
	// 闭区间 [low, high] 升序遍历:对所有满足 IsLessThan()(low, x) <= 0
	// && IsLessThan()(x, high) <= 0 的元素(即 low ≤ x ≤ high)执行 callback,升序触发。
	//  - 调用方需保证 low ≤ high(等价于 IsLessThan()(high, low) <= 0),否则行为未定义
	//  - low == high 时,若树中存在该元素,会触发一次回调
	//  - 大小对比基于 IsLessThan
	//
	// 示例场景：
	// 1. 查询触发时间介于现在、一小时之后之内所有定时任务；
	// 2. 筛选年龄 18‑30 的全部用户；
	// 3. 数据库索引范围查询。
	RangeVisit(low, high T, callback func(T))

	// Rank
	// 返回有多少个元素严格小于入参 item (0‑based 排名前置计数)
	// 例如最小元素 rank=0；第二小 rank=1
	// 严格小于判定:IsLessThan()(x, item) == -1
	//
	// 示例场景：
	// 1. 查询玩家在排行榜上面的名次；
	// 2. 判断一个任务在任务队列当中排在第几位；
	// 3. 顺序统计、分页计算偏移。
	Rank(item T) int

	// Select
	// 根据从零开始的 rank 下标取出对应位置元素；rank 不可大于等于 Size()
	// rank=0 返回最小值
	//
	// 示例场景：
	// 1. 获取排行榜第 5 名用户；
	// 2. 取出前 N 个优先级最高的任务；
	// 3. 实现 k‑th 元素算法、中位数求取 Select(Size()/2)。
	Select(rank int) (T, bool)

	// IsLessThan
	// 返回本树实例绑定的「三态」比较函数:
	//   cmp(a, b) == -1 ⇔ a < b
	//   cmp(a, b) ==  0 ⇔ a == b
	//   cmp(a, b) ==  1 ⇔ a > b
	//
	// 同一棵树的所有方法(Get / Delete / Predecessor / Successor / RangeVisit /
	// Rank / InOrderTraverse / ...)都使用此函数定义全序关系,实现内部不得在
	// 树生命周期内替换该函数;比较器会由 tree 在创建节点时透传到每个 Node,
	// Node.Compare(a, b) 与 IsLessThan()(a, b) 必须返回一致。
	//
	// 调用方可以复用返回的函数对外部数据执行「与树一致」的排序比较。
	// 多个平衡树实例之间若想保证比较器一致,可让它们共用同一个 IsLessThan。
	//
	// 契约:返回结果必须是 -1 / 0 / 1 三者之一(不可返回 2、不可返回 bool),
	// 传入的 IsLessThan 须满足全序关系(反对称、传递、完备)。
	//
	// 示例场景：
	// 1. 外部代码需要按照和平衡树一模一样的逻辑对比两个结构体；
	// 2. 多个平衡树之间保证比较器统一。
	IsLessThan() func(a, b T) int

	// DrawTree
	// 把整棵树以带缩进的树形文本输出到 out（先序：父 → 左 → 右），仅打印节点值。
	// 用于观察树的结构；不输出实现特有的附加字段（如 priority/color）。
	// 空树不输出任何内容。
	//
	// 示例场景：
	// 1. 调试平衡树插入/删除后的结构；
	// 2. demo 中每步操作后打印整棵树。
	DrawTree(out io.Writer)
}
