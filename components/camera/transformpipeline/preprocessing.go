package transformpipeline

import (
	"context"
	"image"

	"github.com/edaniels/gostream"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"

	"go.viam.com/rdk/components/camera"
	"go.viam.com/rdk/rimage"
	"go.viam.com/rdk/rimage/transform"
)

// preprocessDepthTransform applies pre-processing functions to depth maps in order to smooth edges and fill holes.
type preprocessDepthTransform struct {
	stream gostream.VideoStream
}

func newDepthPreprocessTransform(ctx context.Context, source gostream.VideoSource,
) (gostream.VideoSource, camera.ImageType, error) {
	reader := &preprocessDepthTransform{gostream.NewEmbeddedVideoStream(source)}
	var cameraModel *transform.PinholeCameraModel
	if cameraSrc, ok := source.(camera.Camera); ok {
		props, err := cameraSrc.Properties(ctx)
		if err != nil {
			return nil, camera.UnspecifiedStream, err
		}
		cameraModel = &transform.PinholeCameraModel{props.IntrinsicParams, props.DistortionParams}
	}
	cam, err := camera.NewFromReader(ctx, reader, cameraModel, camera.DepthStream)
	return cam, camera.DepthStream, err
}

// Next applies depth preprocessing to the next image.
func (os *preprocessDepthTransform) Read(ctx context.Context) (image.Image, func(), error) {
	ctx, span := trace.StartSpan(ctx, "camera::transformpipeline::depthPreprocess::Read")
	defer span.End()
	i, release, err := os.stream.Next(ctx)
	if err != nil {
		return nil, nil, err
	}
	dm, err := rimage.ConvertImageToDepthMap(ctx, i)
	if err != nil {
		return nil, nil, errors.Wrap(err, "transform source does not provide depth image")
	}
	dm, err = rimage.PreprocessDepthMap(dm, nil)
	if err != nil {
		return nil, nil, err
	}
	return dm, release, nil
}

func (os *preprocessDepthTransform) Close(ctx context.Context) error {
	return os.stream.Close(ctx)
}
