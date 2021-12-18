data maybe(a)
    | Nothing
    | Just(a)

data list(a)
    | Cons(a, list)
    | Nil

data tree(a)
    | Branch(tree, tree)
    | Leaf(a)