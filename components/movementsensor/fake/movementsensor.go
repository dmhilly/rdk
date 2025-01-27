// Package fake is a fake MovementSensor for testing
package fake

import (
	"context"

	"github.com/edaniels/golog"
	"github.com/golang/geo/r3"
	geo "github.com/kellydunn/golang-geo"

	"go.viam.com/rdk/components/movementsensor"
	"go.viam.com/rdk/config"
	"go.viam.com/rdk/registry"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/spatialmath"
)

var modelname = resource.NewDefaultModel("fake")

// AttrConfig is used for converting fake movementsensor attributes.
type AttrConfig struct {
	ConnectionType string `json:"connection_type,omitempty"`
}

func init() {
	registry.RegisterComponent(
		movementsensor.Subtype,
		modelname,
		registry.Component{Constructor: func(
			ctx context.Context,
			deps registry.Dependencies,
			cfg config.Component,
			logger golog.Logger,
		) (interface{}, error) {
			return movementsensor.MovementSensor(&MovementSensor{}), nil
		}})

	config.RegisterComponentAttributeMapConverter(movementsensor.Subtype, modelname,
		func(attributes config.AttributeMap) (interface{}, error) {
			var attr AttrConfig
			return config.TransformAttributeMapToStruct(&attr, attributes)
		},
		&AttrConfig{})
}

// MovementSensor implements is a fake movement sensor interface.
type MovementSensor struct {
	CancelCtx context.Context
	Logger    golog.Logger
	counter   float64
}

// Position gets the position of a fake movementsensor.
func (f *MovementSensor) Position(ctx context.Context, extra map[string]interface{}) (*geo.Point, float64, error) {
	a := f.counter + 50.5
	p := geo.NewPoint(40.7+f.counter, -73.98+f.counter)
	return p, a, nil
}

// LinearVelocity gets the linear velocity of a fake movementsensor.
func (f *MovementSensor) LinearVelocity(ctx context.Context, extra map[string]interface{}) (r3.Vector, error) {
	return r3.Vector{Y: 5.4 + f.counter}, nil
}

// LinearAcceleration gets the linear acceleration of a fake movementsensor.
func (f *MovementSensor) LinearAcceleration(ctx context.Context, extra map[string]interface{}) (r3.Vector, error) {
	return r3.Vector{X: 2.2, Y: 4.5, Z: 2}, nil
}

// AngularVelocity gets the angular velocity of a fake movementsensor.
func (f *MovementSensor) AngularVelocity(ctx context.Context, extra map[string]interface{}) (spatialmath.AngularVelocity, error) {
	return spatialmath.AngularVelocity{Z: 1}, nil
}

// CompassHeading gets the compass headings of a fake movementsensor.
func (f *MovementSensor) CompassHeading(ctx context.Context, extra map[string]interface{}) (float64, error) {
	return 25 + f.counter, nil
}

// Orientation gets the orientation of a fake movementsensor.
func (f *MovementSensor) Orientation(ctx context.Context, extra map[string]interface{}) (spatialmath.Orientation, error) {
	ori := spatialmath.NewEulerAngles()
	ori.Roll = 1 + f.counter
	ori.Pitch = 2 + f.counter
	ori.Yaw = 3 + f.counter
	return ori, nil
}

// DoCommand uses a map string to run custom functionality of a fake movementsensor.
func (f *MovementSensor) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

// Accuracy gets the accuracy of a fake movementsensor.
func (f *MovementSensor) Accuracy(ctx context.Context, extra map[string]interface{}) (map[string]float32, error) {
	return map[string]float32{}, nil
}

// Readings gets the readings of a fake movementsensor.
func (f *MovementSensor) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	return movementsensor.Readings(ctx, f, extra)
}

// Properties returns the properties of a fake movementsensor.
func (f *MovementSensor) Properties(ctx context.Context, extra map[string]interface{}) (*movementsensor.Properties, error) {
	return &movementsensor.Properties{
		LinearVelocitySupported:     true,
		AngularVelocitySupported:    true,
		OrientationSupported:        true,
		PositionSupported:           true,
		CompassHeadingSupported:     true,
		LinearAccelerationSupported: true,
	}, nil
}

// Start returns the fix of a fake gps movementsensor.
func (f *MovementSensor) Start(ctx context.Context) error {
	f.counter = 4.5
	return nil
}

// Close returns the fix of a fake gps movementsensor.
func (f *MovementSensor) Close() error { return nil }

// ReadFix returns the fix of a fake gps movementsensor.
func (f *MovementSensor) ReadFix(ctx context.Context) (int, error) { return 1, nil }
