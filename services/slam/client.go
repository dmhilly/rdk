// Package slam implements simultaneous localization and mapping
// This is an Experimental package
package slam

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"

	"github.com/edaniels/golog"
	"go.opencensus.io/trace"
	pb "go.viam.com/api/service/slam/v1"
	"go.viam.com/utils/protoutils"
	"go.viam.com/utils/rpc"

	"go.viam.com/rdk/pointcloud"
	"go.viam.com/rdk/referenceframe"
	"go.viam.com/rdk/utils"
	"go.viam.com/rdk/vision"
)

// client implements SLAMServiceClient.
type client struct {
	name   string
	conn   rpc.ClientConn
	client pb.SLAMServiceClient
	logger golog.Logger
}

// NewClientFromConn constructs a new Client from the connection passed in.
func NewClientFromConn(ctx context.Context, conn rpc.ClientConn, name string, logger golog.Logger) Service {
	grpcClient := pb.NewSLAMServiceClient(conn)
	c := &client{
		name:   name,
		conn:   conn,
		client: grpcClient,
		logger: logger,
	}
	return c
}

// Position creates a request, calls the slam service Position, and parses the response into the desired PoseInFrame.
func (c *client) Position(ctx context.Context, name string, extra map[string]interface{}) (*referenceframe.PoseInFrame, error) {
	ctx, span := trace.StartSpan(ctx, "slam::client::Position")
	defer span.End()

	ext, err := protoutils.StructToStructPb(extra)
	if err != nil {
		return nil, err
	}

	req := &pb.GetPositionRequest{
		Name:  name,
		Extra: ext,
	}

	resp, err := c.client.GetPosition(ctx, req)
	if err != nil {
		return nil, err
	}
	p := resp.GetPose()
	return referenceframe.ProtobufToPoseInFrame(p), nil
}

// GetMap creates a request, calls the slam service GetMap, and parses the response into the desired mimeType and map data.
func (c *client) GetMap(
	ctx context.Context,
	name, mimeType string,
	cameraPosition *referenceframe.PoseInFrame,
	includeRobotMarker bool,
	extra map[string]interface{},
) (
	string, image.Image, *vision.Object, error,
) {
	ctx, span := trace.StartSpan(ctx, "slam::client::GetMap")
	defer span.End()

	ext, err := protoutils.StructToStructPb(extra)
	if err != nil {
		return "", nil, nil, err
	}

	req := &pb.GetMapRequest{
		Name:               name,
		MimeType:           mimeType,
		IncludeRobotMarker: includeRobotMarker,
		Extra:              ext,
	}

	if cameraPosition != nil {
		req.CameraPosition = referenceframe.PoseInFrameToProtobuf(cameraPosition).Pose
	}

	var imageData image.Image
	vObject := &vision.Object{}

	resp, err := c.client.GetMap(ctx, req)
	if err != nil {
		return "", imageData, vObject, err
	}

	mimeType = resp.MimeType

	switch mimeType {
	case utils.MimeTypeJPEG:
		_, spanDecode := trace.StartSpan(ctx, "slam::client::GetMap::Decode")
		defer spanDecode.End()

		imData := resp.GetImage()
		imageData, err = jpeg.Decode(bytes.NewReader(imData))
		if err != nil {
			return "", imageData, vObject, err
		}
	case utils.MimeTypePCD:
		_, spanGetPC := trace.StartSpan(ctx, "slam::client::GetMap::GetPointCloud")
		defer spanGetPC.End()

		pcData := resp.GetPointCloud()
		pc, err := pointcloud.ReadPCD(bytes.NewReader(pcData.PointCloud))
		if err != nil {
			return "", imageData, vObject, err
		}
		vObject, err = vision.NewObject(pc)
		if err != nil {
			return "", imageData, vObject, err
		}
	}

	return mimeType, imageData, vObject, nil
}
