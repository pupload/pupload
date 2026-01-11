package telemetry

type ExporterType string

const (
	ExporterOTLP   ExporterType = "otlp"
	ExporterStdout ExporterType = "stdout"
	ExporterNone   ExporterType = "none"
)

type TelemetrySettings struct {
	Enabled  bool
	Exporter ExporterType
	Endpoint string

	Insecure bool
	Headers  map[string]string

	SampleRate float64
}
