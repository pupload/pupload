package models

type MimeType string

const (
	// Video
	MimeVideoMP4  MimeType = "video/mp4"
	MimeVideoMKV  MimeType = "video/x-matroska"
	MimeVideoFLV  MimeType = "video/x-flv"
	MimeVideoMPEG MimeType = "video/mpeg"
	MimeVideo3GP  MimeType = "video/3gpp"
	MimeVideoTS   MimeType = "video/mp2t"
	MimeVideoOGG  MimeType = "video/ogg"
	MimeVideoWild MimeType = "video/*"

	// Audio
	MimeAudioFLAC MimeType = "audio/flac"
	MimeAudioOGG  MimeType = "audio/ogg"
	MimeAudioM4A  MimeType = "audio/mp4"
	MimeAudioWMA  MimeType = "audio/x-ms-wma"
	MimeAudioWild MimeType = "audio/*"

	// Image
	MimeImagePNG  MimeType = "image/png"
	MimeImageJPEG MimeType = "image/jpeg"
	MimeImageWebP MimeType = "image/webp"
	MimeImageSVG  MimeType = "image/svg+xml"
	MimeImageBMP  MimeType = "image/bmp"
	MimeImageTIFF MimeType = "image/tiff"
	MimeImageHEIC MimeType = "image/heic"
	MimeImageWild MimeType = "image/*"
)
