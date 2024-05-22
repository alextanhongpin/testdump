package mysqldump

type Transformer func(*SQL) error
