package utils

import "path"

const ContextType = "Content-Type"

func GetContextType(fileName string) string {
	fileExt := path.Ext(fileName)
	// 设置适当的 Content-Type 响应头
	switch fileExt {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".html":
		return "text/html"
	case ".txt":
		return "text/plain"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	default:
		return "application/octet-stream"
	}
}
