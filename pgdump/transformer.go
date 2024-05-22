package pgdump

type Transformer func(*SQL) error
