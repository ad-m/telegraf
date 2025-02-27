package shim

import (
	"bufio"
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
)

// AddOutput adds the input to the shim. Later calls to Run() will run this.
func (s *Shim) AddOutput(output telegraf.Output) error {
	models.SetLoggerOnPlugin(output, s.Log())
	if p, ok := output.(telegraf.Initializer); ok {
		err := p.Init()
		if err != nil {
			return fmt.Errorf("failed to init input: %w", err)
		}
	}

	s.Output = output
	return nil
}

func (s *Shim) RunOutput() error {
	parser := influx.Parser{}
	err := parser.Init()
	if err != nil {
		return fmt.Errorf("failed to create new parser: %w", err)
	}

	err = s.Output.Connect()
	if err != nil {
		return fmt.Errorf("failed to start processor: %w", err)
	}
	defer s.Output.Close()

	var m telegraf.Metric

	scanner := bufio.NewScanner(s.stdin)
	for scanner.Scan() {
		m, err = parser.ParseLine(scanner.Text())
		if err != nil {
			fmt.Fprintf(s.stderr, "Failed to parse metric: %s\n", err)
			continue
		}
		if err = s.Output.Write([]telegraf.Metric{m}); err != nil {
			fmt.Fprintf(s.stderr, "Failed to write metric: %s\n", err)
		}
	}

	return nil
}
