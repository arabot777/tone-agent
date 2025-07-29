package utils

import "crypto/md5"

func ComputeMd5Checksum(data []byte) []byte {
	h := md5.New()
	h.Write(data)
	return h.Sum(nil)
}
