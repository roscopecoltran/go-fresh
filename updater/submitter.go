package updater

type Submitter interface {
	SubmitPR() error
}

func SubmitPR(project string, from, to string) error {
	submitter, err := newNomadSubmitter()
	if err != nil {
		return err
	}
	return submitter.SubmitPR()
}
