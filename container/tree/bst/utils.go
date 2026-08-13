package bst

import (
	"fmt"
	"io"
	"reflect"
)

// nodeIsNil 判断 p 是否为空：既包括 nil 接口，也包括「typed nil」接口
// （接口类型非 nil、但其包裹的具体指针为 nil，如 *treapNode(nil) 作为
// BSTNodeInterface 传入）。后者用 p == nil 无法识别，会引发 nil 指针解引用。
func nodeIsNil[T any](p BSTNodeInterface[T]) bool {
	if p == nil {
		return true
	}
	rv := reflect.ValueOf(p)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	}
	return false
}

// 工具:判断 child 是否为 parent 的左孩子
func IsLeftChild[T any](parent, child BSTNodeInterface[T]) bool {
	if parent == nil {
		return false
	}
	return parent.GetLeft() == child
}

// UpdateSize 重新计算 n 所在子树的 size 并写回 n 自身。
// 约定:size(n) = 1 + size(n.GetLeft().GetSize()) + size(n.GetRight().GetSize());空子树 size = 0。
// n 为 nil 时直接返回,不会触发任何 SetSize。
//
// 典型用法:旋转(RotateLeft / RotateRight)完成后,沿旋转路径
// 自底向上对每个相关节点调用本函数,即可保证整条链路的 size 不变量。
//
// 注意:本函数只刷新 n 自身的 size,**不会**沿 Parent 链向上回溯;
//
//	旋转路径上「n 的祖先节点」的 size 必须由调用方显式刷新
//	(因为 rotate 函数本身只更新旋转局部的 size,见各 rotate 文档)。
func UpdateSize[T any](n BSTNodeInterface[T]) {
	if n == nil {
		return
	}
	sz := 1
	if l := n.GetLeft(); l != nil {
		sz += l.GetSize()
	}
	if r := n.GetRight(); r != nil {
		sz += r.GetSize()
	}
	n.SetSize(sz)
}

// RotateLeft 对以 p 为根的子树做左旋。
// 要求:p 必须存在,且 p.GetRight() != nil(否则直接返回 p,无副作用)。
// 返回值:旋转之后该子树的「新根」(即原 p.GetRight(),记作 r)。
//
// 左旋不变量(局部):
//   - p 的新右孩子 = 原 r.GetLeft() (记作 rl)
//   - r 的新左孩子 = p
//   - r 接管 p 原来的父亲 pp(若 pp != nil)的对应孩子位;
//     若 pp == nil,则 r 成为新的整棵树的根(Parent = nil)
//   - 所有相关节点(p / r / rl / pp)的 Parent 指针均被同步更新
//
// 不变量(size):
//   - 本函数只刷新 p 和 r 自身的 size(且先 p 后 r,顺序依赖)
//   - pp 及其祖先的 size **不会**自动刷新,由调用方沿父链自底向上重算
//
// 注意:rotate 是纯结构操作,与比较器无关;不需要 node.Compare。
//
// 图示(→ 表示父指针):
//
//	 pp              pp
//	  |               |
//	  p               r
//	 / \             / \
//	L   r    →     p   R
//	   / \         / \
//	  rl  R       L   rl
func RotateLeft[T any](p BSTNodeInterface[T]) BSTNodeInterface[T] {
	r := p.GetRight()
	if r == nil {
		return p
	}
	rl := r.GetLeft()
	pp := p.GetParent()

	// 1. p 的右孩子接管 r 的左子结点
	p.SetRight(rl)
	if rl != nil {
		rl.SetParent(p)
	}

	// 2. r 向下挂载 p
	r.SetLeft(p)
	p.SetParent(r)

	//3. 将 r 接入原先的上层父节点 pp
	if pp != nil {
		if IsLeftChild(pp, p) {
			pp.SetLeft(r)
		} else {
			pp.SetRight(r)
		}
	}
	r.SetParent(pp)

	//4. 自底向上刷新 size
	UpdateSize(p)
	UpdateSize(r)

	return r
}

// RotateRight 是 RotateLeft 的镜像:对以 p 为根的子树做右旋。
// 要求:p 必须存在,且 p.GetLeft() != nil(否则直接返回 p,无副作用)。
// 返回值:旋转之后该子树的「新根」(即原 p.GetLeft(),记作 l)。
//
// 右旋不变量(局部):
//   - p 的新左孩子 = 原 l.GetRight() (记作 lr)
//   - l 的新右孩子 = p
//   - l 接管 p 原来的父亲 pp(若 pp != nil)的对应孩子位;
//     若 pp == nil,则 l 成为新的整棵树的根(Parent = nil)
//   - 所有相关节点(p / l / lr / pp)的 Parent 指针均被同步更新
//
// 不变量(size):
//   - 本函数只刷新 p 和 l 自身的 size(且先 p 后 l,顺序依赖)
//   - pp 及其祖先的 size **不会**自动刷新,由调用方沿父链自底向上重算
//
// 注意:rotate 是纯结构操作,与比较器无关;不需要 node.Compare。
//
// 图示(→ 表示父指针):
//
//	   pp              pp
//	    |               |
//	    p               l
//	   / \             / \
//	  l   R    →     L   p
//	 / \                 / \
//	L  lr               lr  R
func RotateRight[T any](p BSTNodeInterface[T]) BSTNodeInterface[T] {
	l := p.GetLeft()
	if l == nil {
		return p
	}
	lr := l.GetRight()
	pp := p.GetParent()

	p.SetLeft(lr)
	if lr != nil {
		lr.SetParent(p)
	}

	l.SetRight(p)
	p.SetParent(l)

	if pp != nil {
		if IsLeftChild(pp, p) {
			pp.SetLeft(l)
		} else {
			pp.SetRight(l)
		}
	}
	l.SetParent(pp)

	UpdateSize(p)
	UpdateSize(l)
	return l
}

func Get[T any](p BSTNodeInterface[T], item T) (BSTNodeInterface[T], bool) {
	if nodeIsNil(p) {
		return nil, false
	}

	cmpFn := p.CompareFn()
	if cmpFn(p.GetVal(), item) == 0 {
		return p, true
	}

	if cmpFn(p.GetVal(), item) < 0 {
		return Get(p.GetRight(), item)
	} else {
		return Get(p.GetLeft(), item)
	}
}

func Insert[T any](p *Node[T], pp *Node[T], left bool, item T, cmp CMPFN[T]) *Node[T] {
	if p == nil {
		ret := NewNode(item, cmp)
		if pp == nil {
			return ret
		}
		if left {
			pp.SetLeft(ret)
		} else {
			pp.SetRight(ret)
		}
		ret.SetParent(pp)
		RefreshSizeUp(pp)
		return ret
	}

	if cmp(p.val, item) >= 0 {
		return Insert(p.leftNode(), p, true, item, cmp)
	}
	return Insert(p.rightNode(), p, false, item, cmp)
}

// RefreshSizeUp 自 n 起沿父链向上逐个调用 bst.UpdateSize,回填子树 size。
func RefreshSizeUp[T any](n BSTNodeInterface[T]) {
	for n != nil {
		UpdateSize(n)
		n = n.GetParent()
	}
}

// InOrderTraverse 中序遍历以 p 为根的子树，按节点比较器升序对每个元素调用 fn。
func InOrderTraverse[T any](p BSTNodeInterface[T], fn func(T)) {
	if nodeIsNil(p) {
		return
	}
	InOrderTraverse(p.GetLeft(), fn)
	fn(p.GetVal())
	InOrderTraverse(p.GetRight(), fn)
}

// PreorderTraverse 前序遍历以 p 为根的子树：先访问当前节点，再左、再右。
func PreorderTraverse[T any](p BSTNodeInterface[T], fn func(T)) {
	if nodeIsNil(p) {
		return
	}
	fn(p.GetVal())
	PreorderTraverse(p.GetLeft(), fn)
	PreorderTraverse(p.GetRight(), fn)
}

// PostorderTraverse 后序遍历以 p 为根的子树：先左、再右，最后访问当前节点。
func PostorderTraverse[T any](p BSTNodeInterface[T], fn func(T)) {
	if nodeIsNil(p) {
		return
	}
	PostorderTraverse(p.GetLeft(), fn)
	PostorderTraverse(p.GetRight(), fn)
	fn(p.GetVal())
}

// Min 返回以 p 为根子树的最小元素；p 为 nil 时返回 (zero, false)。
func Min[T any](p BSTNodeInterface[T]) (T, bool) {
	if nodeIsNil(p) {
		var zero T
		return zero, false
	}
	for p.GetLeft() != nil {
		p = p.GetLeft()
	}
	return p.GetVal(), true
}

// Max 返回以 p 为根子树的最大元素；p 为 nil 时返回 (zero, false)。
func Max[T any](p BSTNodeInterface[T]) (T, bool) {
	if nodeIsNil(p) {
		var zero T
		return zero, false
	}
	for p.GetRight() != nil {
		p = p.GetRight()
	}
	return p.GetVal(), true
}

// Predecessor 返回以 p 为根的子树中严格小于 item 的最大元素。
// p 为 nil 或不存在这样的前驱时返回 (zero, false)。
func Predecessor[T any](p BSTNodeInterface[T], item T) (T, bool) {
	var zero T
	var pred BSTNodeInterface[T]
	for !nodeIsNil(p) {
		if p.CompareFn()(p.GetVal(), item) < 0 {
			pred = p
			p = p.GetRight()
		} else {
			p = p.GetLeft()
		}
	}
	if pred == nil {
		return zero, false
	}
	return pred.GetVal(), true
}

// Successor 返回以 p 为根的子树中严格大于 item 的最小元素。
// p 为 nil 或不存在这样的后继时返回 (zero, false)。
func Successor[T any](p BSTNodeInterface[T], item T) (T, bool) {
	var zero T
	var succ BSTNodeInterface[T]
	for !nodeIsNil(p) {
		if p.CompareFn()(p.GetVal(), item) > 0 {
			succ = p
			p = p.GetLeft()
		} else {
			p = p.GetRight()
		}
	}
	if succ == nil {
		return zero, false
	}
	return succ.GetVal(), true
}

// Rank 返回以 p 为根的子树中严格小于 item 的元素个数（0-based）。
func Rank[T any](p BSTNodeInterface[T], item T) int {
	var count int
	for !nodeIsNil(p) {
		if p.CompareFn()(p.GetVal(), item) < 0 {
			// 当前节点 < item：计入当前节点 + 其左子树
			count++
			if l := p.GetLeft(); l != nil {
				count += l.GetSize()
			}
			p = p.GetRight()
		} else {
			// 当前节点 >= item：不计入，向左
			p = p.GetLeft()
		}
	}
	return count
}

// Select 返回以 p 为根的子树中第 rank 小（0-based）的元素；rank 越界返回 (zero, false)。
func Select[T any](p BSTNodeInterface[T], rank int) (T, bool) {
	var zero T
	for !nodeIsNil(p) {
		lsz := 0
		if l := p.GetLeft(); l != nil {
			lsz = l.GetSize()
		}
		switch {
		case rank < lsz:
			p = p.GetLeft()
		case rank == lsz:
			return p.GetVal(), true
		default: // rank > lsz
			rank -= lsz + 1
			p = p.GetRight()
		}
	}
	return zero, false
}

// RangeVisit 闭区间 [low, high] 升序遍历以 p 为根的子树：
// 对所有 low ≤ x ≤ high（按节点比较器判定）的元素调用 fn，升序触发。
// 调用方需保证 low ≤ high，否则行为未定义。
func RangeVisit[T any](p BSTNodeInterface[T], low, high T, fn func(T)) {
	if nodeIsNil(p) {
		return
	}
	cmp := p.CompareFn()
	val := p.GetVal()

	// 向左剪枝：仅当 val >= low 时，左子树才可能存在 ≥ low 的节点。
	// 注意用 <= 0 而非 < 0：相等元素允许共存且都挂在左侧，val == low 时
	// 左子树仍可能有等于 low 的重复节点，必须继续向左搜索。
	if cmp(low, val) <= 0 {
		RangeVisit(p.GetLeft(), low, high, fn)
	}

	// 当前节点落在 [low, high] 内才触发
	if cmp(low, val) <= 0 && cmp(val, high) <= 0 {
		fn(val)
	}

	// 向右剪枝：仅当 val <= high 时，右子树才可能存在 ≤ high 的节点。
	// 对称地，val == high 时右子树仍可能有等于 high 的重复节点，需继续向右搜索。
	if cmp(val, high) <= 0 {
		RangeVisit(p.GetRight(), low, high, fn)
	}
}

// IsOrderedNode 检查节点 n 是否仍满足 BST 有序不变量。
// 必须与 BST 有序约定一致：左子树 n >= l（允许相等）、右子树 n < r（严格）。
// 需同时验证 n 与 parent、n 与左孩子、n 与右孩子 三处关系。
func IsOrderedNode[T any](n BSTNodeInterface[T]) bool {
	if nodeIsNil(n) {
		return true
	}
	cmp := n.CompareFn()
	// 1) 与父节点
	if pp := n.GetParent(); pp != nil {
		if IsLeftChild(pp, n) {
			// n 是左孩子：须满足 pp >= n；pp < n 则乱
			if cmp(pp.GetVal(), n.GetVal()) < 0 {
				return false
			}
		} else {
			// n 是右孩子：须满足 n > pp（严格）；n <= pp 则乱
			if cmp(n.GetVal(), pp.GetVal()) <= 0 {
				return false
			}
		}
	}
	// 2) 与左孩子：须满足 n >= l；n < l 则乱
	if l := n.GetLeft(); l != nil {
		if cmp(n.GetVal(), l.GetVal()) < 0 {
			return false
		}
	}
	// 3) 与右孩子：须满足 n < r；n >= r 则乱
	if r := n.GetRight(); r != nil {
		if cmp(n.GetVal(), r.GetVal()) >= 0 {
			return false
		}
	}
	return true
}

// DrawTree 以横向树形图输出以 p 为根的子树：
// 左右子树分列两侧，父节点在上，用 / \ 斜线连接，直观展示树的层级与左右分支。
// 仅打印节点值，不涉及任何实现特有的附加字段（如 priority/color）。
//
// 输出示例：
//
//	     50
//	   /    \
//	  30     70
//	 /  \   /  \
//	20  40 60  80
func DrawTree[T any](p BSTNodeInterface[T], out io.Writer) {
	if nodeIsNil(p) {
		return
	}
	lines, rootIdx := renderSub(p)
	// 把整棵树平移，使根文本居中于输出宽度，避免左分支被挤到边缘。
	totalW := maxLineLen(lines)
	pad := (totalW - (2*rootIdx + 1)) / 2 // 让根中心接近 totalW/2
	if pad < 2 {
		pad = 2 // 至少保留左边距
	}
	for _, l := range lines {
		fmt.Fprintln(out, padRight(spaces(pad)+l, totalW+pad))
	}
}

// renderSub 返回以 n 为根子树的横向文本行（每行等宽）以及根文本中心所在列。
// 父节点文本中心放在左、右孩子根中心的中点；单孩子时放在孩子根中心。
// 该算法保证每个父节点严格居中于左右分支之间。
func renderSub[T any](n BSTNodeInterface[T]) ([]string, int) {
	val := fmt.Sprint(n.GetVal())
	vlen := len(val)
	l := n.GetLeft()
	r := n.GetRight()
	hasL := !nodeIsNil(l)
	hasR := !nodeIsNil(r)

	if !hasL && !hasR {
		line := spaces(vlen/2) + val // 根居中，行宽 vlen
		return []string{line}, vlen / 2
	}

	// 递归渲染左右子树
	var leftLines, rightLines []string
	var leftIdx, rightIdx int
	if hasL {
		leftLines, leftIdx = renderSub(l)
	}
	if hasR {
		rightLines, rightIdx = renderSub(r)
	}

	leftW := 0
	if hasL {
		leftW = maxLineLen(leftLines)
	}
	rightW := 0
	if hasR {
		rightW = maxLineLen(rightLines)
	}

	gap := 2 // 左右子树最小间隔（加大以拉开水平跨度，斜线更清晰）
	// 左右子树根中心在拼接后坐标系的绝对列
	absL := leftIdx
	absR := leftW + gap + rightIdx
	totalW := leftW + gap + rightW

	// 父中心：
	//  - 双孩子：取两中心中点（父居中于两分支）
	//  - 单左孩子：父偏向右侧，使左分支斜向左下
	//  - 单右孩子：父偏向左侧，使右分支斜向右下
	rootIdx := absL
	switch {
	case hasL && hasR:
		rootIdx = (absL + absR) / 2
	case hasL:
		// 单左孩子：父在左孩子右上方，偏移半个 gap 以拉出斜线
		rootIdx = absL + gap
	case hasR:
		// 单右孩子：父在右孩子左上方
		rootIdx = absR - gap
	}

	// 父文本起点，使中心落在 rootIdx
	mid := rootIdx - vlen/2
	if mid < 0 {
		mid = 0
	}
	// 保证总宽容纳父文本
	if m := mid + vlen; m > totalW {
		totalW = m
	}

	// 拼接 body
	rows := maxLen(len(leftLines), len(rightLines))
	body := make([]string, 0, rows+2)
	for i := 0; i < rows; i++ {
		var lseg, rseg string
		if hasL && i < len(leftLines) {
			lseg = leftLines[i]
		}
		if hasR && i < len(rightLines) {
			rseg = rightLines[i]
		}
		body = append(body, padRight(lseg, leftW)+spaces(gap)+padRight(rseg, rightW))
	}

	// 连接行：仅对实际存在的孩子画斜线
	conn := make([]byte, totalW)
	for i := range conn {
		conn[i] = ' '
	}
	if hasL && absL >= 0 && absL < totalW {
		conn[absL] = '/'
	}
	if hasR && absR >= 0 && absR < totalW {
		conn[absR] = '\\'
	}

	parentLine := padRight(spaces(mid)+val, totalW)
	connect := []string{parentLine, string(conn)}

	out := append(connect, body...)
	return out, rootIdx
}

// maxLineLen 返回 lines 中最长一行的字节长度。
func maxLineLen(lines []string) int {
	m := 0
	for _, l := range lines {
		if len(l) > m {
			m = len(l)
		}
	}
	return m
}

func maxLen(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func padRight(s string, w int) string {
	for len(s) < w {
		s += " "
	}
	return s
}

func spaces(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = ' '
	}
	return string(b)
}
