package models

type DataWell struct {
	Store string
	Edge  string
	Type  string  // dynamic or static
	Key   *string // defaults to artifact id. ${RUN_ID}, ${ARTIFACT_ID}, ${NODE_ID}

}
