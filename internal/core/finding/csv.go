package finding

import (
	"encoding/csv"
	"os"
)

func getCSVHeader() []string {
	return []string{
		"FilePath",
		"Line Number",
		"Action",
		"Description",
		"SignatureID",
		"Finding List",
		"Repo Owner",
		"Repo Name",
		"Commit Hash",
		"Commit Message",
		"Commit Author",
		"File URL",
		"Secret ID",
		"App Version",
		"Signatures Version",
	}
}

func (f *Finding) toCSV() []string {
	return []string{
		f.FilePath,
		f.LineNumber,
		f.Action,
		f.Description,
		f.SignatureID,
		f.Content,
		f.RepositoryOwner,
		f.RepositoryName,
		f.CommitHash,
		f.CommitMessage,
		f.CommitAuthor,
		f.FileURL,
		f.SecretID,
		f.AppVersion,
		f.SignatureVersion,
	}
}

func WriteCSV(findings []*Finding) error {
	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	err := w.Write(getCSVHeader())
	if err != nil {
		return err
	}

	for _, v := range findings {
		err := w.Write(v.toCSV())
		if err != nil {
			return err
		}
	}

	return nil
}
