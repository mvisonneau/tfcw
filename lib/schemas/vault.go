package schemas

type Vault struct {
	Address *string            `hcl:"address"`
	Token   *string            `hcl:"token"`
	Method  *string            `hcl:"method"`
	Params  *map[string]string `hcl:"params"`
	Path    *string            `hcl:"path"`
	Key     *string            `hcl:"key"`
	Keys    *map[string]string `hcl:"keys"`
}
