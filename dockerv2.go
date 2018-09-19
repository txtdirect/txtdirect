package txtdirect

func generateDockerv2URI(path string, rec record) (string, int) {
	// TODO: parse path in future to support all docker apis
	return rec.To, rec.Code
}
