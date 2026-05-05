package oauth

type OAuth struct {
	StateManager   *StateManager
	GithubProvider Provider
	GoogleProvider Provider
}
