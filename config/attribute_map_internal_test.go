package config_test

import (
	"fmt"
	"testing"

	"go.viam.com/rdk/config"
	"go.viam.com/rdk/robot/packages"
	"go.viam.com/rdk/services/vision"
)

// actually this is external testing... anyways

var modelPath string = "${packages.i_am_a_model}/mobilenet_v1_1.0_224_quant.tflite"
var labelPath string = "${packages.i_am_a_model}/lorem.txt"

type MyStruct struct {
	A *int
	B *string
	C *map[int]string
}

func TestAttributeWalker(t *testing.T) {
	visionAttributes := &vision.Attributes{
		ModelRegistry: []vision.VisModelConfig{
			{
				Name: "my_classifier",
				Type: "classifications",
				Parameters: config.AttributeMap(map[string]interface{}{
					"num_threads": 1,
					"model_path":  modelPath,
					"label_path":  labelPath,
				}),
			},
		},
	}

	newAttr, err := config.WalkAttr(visionAttributes, packages.NewPackagePathVisitor(packages.NewNoopManager()))
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(visionAttributes)
	fmt.Println(newAttr)

	i := 17
	s := &MyStruct{
		A: &i,
		B: &modelPath,
		C: &map[int]string{
			1: labelPath,
			4: labelPath,
			7: modelPath,
		},
	}

	newAttr, err = config.WalkAttr(s, packages.NewPackagePathVisitor(packages.NewNoopManager()))
	if err != nil {
		t.Fatal(err)
	}
}
