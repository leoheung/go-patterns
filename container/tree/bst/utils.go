package bst

// 工具:判断 child 是否为 parent 的左孩子
func IsLeftChild[T any](parent, child BSTNodeInterface[T]) bool {
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
