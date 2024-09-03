package upload

/*
func (s *Server) uploadUserFiles(ctx context.Context, userId int, files map[string]*multipart.FileHeader) error {
	infos := make([]*net.UploadFileInfo, 0, len(files))

	for path, file := range files {
		info, err := readFile(path, file)
		if err != nil {
			return err
		}
		infos = append(infos, info)
	}

	_, err := net.HttpPostStatic(s.logger, s.global.Settings.StaticStorage.Url, "image", infos)
	return err
}

func readFile(path string, file *multipart.FileHeader) (*net.UploadFileInfo, error) {

	f, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data := make([]byte, file.Size, file.Size)
	_, err = f.Read(data)

	if err != nil {
		return nil, err
	}

	info := &net.UploadFileInfo{Path: path, Data: data}

	return info, nil
}
*/
