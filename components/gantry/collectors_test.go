package gantry_test

import (
	"context"
	"testing"
	"time"

	clk "github.com/benbjohnson/clock"
	"github.com/edaniels/golog"
	pb "go.viam.com/api/component/gantry/v1"
	"go.viam.com/test"

	"go.viam.com/rdk/components/gantry"
	"go.viam.com/rdk/data"
	tu "go.viam.com/rdk/testutils"
	"go.viam.com/rdk/testutils/inject"
)

const (
	componentName   = "gantry"
	captureInterval = time.Second
)

var floatList = []float64{1.0, 2.0, 3.0}

func TestGantryCollectors(t *testing.T) {
	tests := []struct {
		name      string
		collector data.CollectorConstructor
		expected  map[string]any
	}{
		{
			name:      "Length collector should write a lengths response",
			collector: gantry.NewLengthsCollector,
			expected: tu.ToProtoMapIgnoreOmitEmpty(pb.GetLengthsResponse{
				LengthsMm: scaleMetersToMm(floatList),
			}),
		},
		{
			name:      "Position collector should write a list of positions",
			collector: gantry.NewPositionCollector,
			expected: tu.ToProtoMapIgnoreOmitEmpty(pb.GetPositionResponse{
				PositionsMm: scaleMetersToMm(floatList),
			}),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockClock := clk.NewMock()
			buf := tu.MockBuffer{}
			params := data.CollectorParams{
				ComponentName: componentName,
				Interval:      captureInterval,
				Logger:        golog.NewTestLogger(t),
				Clock:         mockClock,
				Target:        &buf,
			}

			gantry := newGantry()
			col, err := tc.collector(gantry, params)
			test.That(t, err, test.ShouldBeNil)

			defer col.Close()
			col.Collect()
			mockClock.Add(captureInterval)

			test.That(t, buf.Length(), test.ShouldEqual, 1)
			test.That(t, buf.Writes[0].GetStruct().AsMap(), test.ShouldResemble, tc.expected)
		})
	}
}

func newGantry() gantry.Gantry {
	g := &inject.Gantry{}
	g.PositionFunc = func(ctx context.Context, extra map[string]interface{}) ([]float64, error) {
		return floatList, nil
	}
	g.LengthsFunc = func(ctx context.Context, extra map[string]interface{}) ([]float64, error) {
		return floatList, nil
	}
	return g
}

func scaleMetersToMm(meters []float64) []float64 {
	ret := make([]float64, len(meters))
	for i := range ret {
		ret[i] = meters[i] * 1000
	}
	return ret
}
