package api

func Shutdown() error {
	if db != nil {
		err := db.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
