package forms

type errors map[string][]string

// Add přidá chybovou zprávu pro dané pole formuláře
func (e errors) Add(field, message string) {
	e[field] = append(e[field], message)
}

// Get vrací první chybovou zprávu
func (e errors) Get(field string) string {
	es := e[field]
	if len(es) == 0 {
		return ""
	}
	return es[0]
}
