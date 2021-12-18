data maybe(a)
    | Nothing
    | Just(a)

data list(a)
    | Cons(a, list)
    | Nil