<script setup lang="ts">

import { grpc } from '@improbable-eng/grpc-web';
import { ref, onMounted, onUnmounted } from 'vue';
import { Client, type ServiceError, baseApi, commonApi, StreamClient } from '@viamrobotics/sdk';
import { filterResources } from '../lib/resource';
import { displayError } from '../lib/error';
import KeyboardInput, { type Keys } from './keyboard-input.vue';
import { rcLogConditionally } from '../lib/log';
import { cameraStreamStates, baseStreamStates } from '../lib/camera-state';

interface Props {
  name: string;
  resources: commonApi.ResourceName.AsObject[];
  client: Client;
}

// eslint-disable-next-line no-shadow
const enum Keymap {
  LEFT = 'a',
  RIGHT = 'd',
  FORWARD = 'w',
  BACKWARD = 's'
}

const props = defineProps<Props>();

type Tabs = 'Keyboard' | 'Discrete'
type MovementTypes = 'Continuous' | 'Discrete'
type MovementModes = 'Straight' | 'Spin'
type SpinTypes = 'Clockwise' | 'Counterclockwise'
type Directions = 'Forwards' | 'Backwards'

const selectedItem = ref<Tabs>('Keyboard');
const movementMode = ref<MovementModes>('Straight');
const movementType = ref<MovementTypes>('Continuous');
const direction = ref<Directions>('Forwards');
const spinType = ref<SpinTypes>('Clockwise');
const increment = ref(1000);
// straight mm/s
const speed = ref(300);
// deg/s
const spinSpeed = ref(90);
const angle = ref(0);

const selectCameras = ref('');

const power = $ref(50);

const pressed = new Set<Keys>();
let stopped = true;

const initStreamState = () => {
  for (const value of filterResources(props.resources, 'rdk', 'component', 'camera')) {
    baseStreamStates.set(value.name, false);
  }
};

const resetDiscreteState = () => {
  movementMode.value = 'Straight';
  movementType.value = 'Continuous';
  direction.value = 'Forwards';
  spinType.value = 'Clockwise';
};

const setMovementMode = (mode: MovementModes) => {
  movementMode.value = mode;
};

const setMovementType = (type: MovementTypes) => {
  movementType.value = type;
};

const setSpinType = (type: SpinTypes) => {
  spinType.value = type;
};

const setDirection = (dir: Directions) => {
  direction.value = dir;
};

const stop = () => {
  stopped = true;
  const req = new baseApi.StopRequest();
  req.setName(props.name);
  rcLogConditionally(req);
  props.client.baseService.stop(req, new grpc.Metadata(), displayError);
};

const digestInput = () => {
  let linearValue = 0;
  let angularValue = 0;

  for (const item of pressed) {
    switch (item) {
      case Keymap.FORWARD: {
        linearValue += Number(0.01 * power);
        break;
      }
      case Keymap.BACKWARD: {
        linearValue -= Number(0.01 * power);
        break;
      }
      case Keymap.LEFT: {
        angularValue += Number(0.01 * power);
        break;
      }
      case Keymap.RIGHT: {
        angularValue -= Number(0.01 * power);
        break;
      }
    }
  }

  const linear = new commonApi.Vector3();
  const angular = new commonApi.Vector3();
  linear.setY(linearValue);
  angular.setZ(angularValue);

  const req = new baseApi.SetPowerRequest();
  req.setName(props.name);
  req.setLinear(linear);
  req.setAngular(angular);

  rcLogConditionally(req);
  props.client.baseService.setPower(req, new grpc.Metadata(), (error) => {
    displayError(error);

    if (pressed.size <= 0) {
      stop();
    }
  });
};

const handleKeyDown = (key: Keys) => {
  pressed.add(key);
  digestInput();
};

const handleKeyUp = (key: Keys) => {
  pressed.delete(key);

  if (pressed.size > 0) {
    stopped = false;
    digestInput();
  } else {
    stop();
  }
};

const handleBaseStraight = (name: string, event: {
  distance: number
  speed: number
  direction: number
  movementType: MovementTypes
}) => {
  if (event.movementType === 'Continuous') {
    const linear = new commonApi.Vector3();
    linear.setY(event.speed * event.direction);

    const req = new baseApi.SetVelocityRequest();
    req.setName(name);
    req.setLinear(linear);
    req.setAngular(new commonApi.Vector3());

    rcLogConditionally(req);
    props.client.baseService.setVelocity(req, new grpc.Metadata(), displayError);
  } else {
    const req = new baseApi.MoveStraightRequest();
    req.setName(name);
    req.setMmPerSec(event.speed * event.direction);
    req.setDistanceMm(event.distance);

    rcLogConditionally(req);
    props.client.baseService.moveStraight(req, new grpc.Metadata(), displayError);
  }
};

const baseRun = () => {
  if (movementMode.value === 'Spin') {

    const req = new baseApi.SpinRequest();
    req.setName(props.name);
    req.setAngleDeg(angle.value * (spinType.value === 'Clockwise' ? -1 : 1));
    req.setDegsPerSec(spinSpeed.value);

    rcLogConditionally(req);
    props.client.baseService.spin(req, new grpc.Metadata(), displayError);

  } else if (movementMode.value === 'Straight') {

    handleBaseStraight(props.name, {
      movementType: movementType.value,
      direction: direction.value === 'Forwards' ? 1 : -1,
      speed: speed.value,
      distance: increment.value,
    });

  }
};

const viewPreviewCamera = (values: string) => {
  const streams = new StreamClient(props.client);
  for (const [key] of baseStreamStates) {
    if (values.split(',').includes(key)) {
      try {
        // Only add stream if other components have not already
        if (!cameraStreamStates.get(key) && !baseStreamStates.get(key)) {
          streams.add(key);
        }
      } catch (error) {
        displayError(error as ServiceError);
      }
      baseStreamStates.set(key, true);
    } else if (baseStreamStates.get(key) === true) {
      try {
        // Only remove stream if other components are not using the stream
        if (!cameraStreamStates.get(key)) {
          streams.remove(key);
        }
      } catch (error) {
        displayError(error as ServiceError);
      }
      baseStreamStates.set(key, false);
    }
  }
};

const handleTabSelect = (tab: Tabs) => {
  selectedItem.value = tab;

  /*
   * deselect options from select cameras select
   * TODO: handle better with xstate and reactivate on return
   */
  selectCameras.value = '';
  viewPreviewCamera(selectCameras.value);

  if (tab === 'Discrete') {
    resetDiscreteState();
  }
};

const handleVisibilityChange = () => {
  if (document.visibilityState === 'hidden') {
    pressed.clear();
    stop();
  }
};

onMounted(() => {
  initStreamState();
  window.addEventListener('visibilitychange', handleVisibilityChange);
});

onUnmounted(() => {
  stop();
  window.removeEventListener('visibilitychange', handleVisibilityChange);
});
</script>

<template>
  <v-collapse
    :title="name"
    class="base"
  >
    <v-breadcrumbs
      slot="title"
      crumbs="base"
    />

    <v-button
      slot="header"
      variant="danger"
      icon="stop-circle"
      label="STOP"
      @click="stop"
    />

    <div class="border border-t-0 border-black pt-2">
      <v-tabs
        tabs="Keyboard, Discrete"
        :selected="selectedItem"
        @input="handleTabSelect($event.detail.value)"
      />

      <div
        v-if="selectedItem === 'Keyboard'"
        class="h-auto p-4"
      >
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-8">
          <div class="flex flex-col gap-4">
            <KeyboardInput
              @keydown="handleKeyDown"
              @keyup="handleKeyUp"
              @toggle="(active: boolean) => { !active && (pressed.size > 0 || !stopped) && stop() }"
            />
            <v-slider
              id="power"
              class="pt-2 w-full max-w-xs"
              :min="0"
              :max="100"
              :step="1"
              suffix="%"
              label="Power %"
              :value="power"
              @input="power = $event.detail.value"
            />
          </div>
          <div v-if="filterResources(resources, 'rdk', 'component', 'camera')">
            <v-multiselect
              v-model="selectCameras"
              class="mb-4"
              clearable="false"
              placeholder="Select Cameras"
              aria-label="Select Cameras"
              :options="
                filterResources(resources, 'rdk', 'component', 'camera')
                  .map(({ name }) => name)
                  .join(',')
              "
              @input="viewPreviewCamera($event.detail.value)"
            />
            <template
              v-for="basecamera in filterResources(
                resources,
                'rdk',
                'component',
                'camera'
              )"
              :key="basecamera.name"
            >
              <div
                v-if="basecamera"
                :data-stream-preview="basecamera.name"
                :class="{ 'hidden': !baseStreamStates.get(basecamera.name) }"
              />
            </template>
          </div>
        </div>
      </div>
      <div
        v-if="selectedItem === 'Discrete'"
        class="flex gap-4 p-4"
      >
        <div class="mb-4 grow">
          <v-radio
            label="Movement Mode"
            options="Straight, Spin"
            :selected="movementMode"
            @input="setMovementMode($event.detail.value)"
          />
          <div class="flex flex-wrap items-center gap-2 pt-4">
            <v-radio
              v-if="movementMode === 'Straight'"
              label="Movement Type"
              options="Continuous, Discrete"
              :selected="movementType"
              @input="setMovementType($event.detail.value)"
            />
            <v-radio
              v-if="movementMode === 'Straight'"
              label="Direction"
              options="Forwards, Backwards"
              :selected="direction"
              @input="setDirection($event.detail.value)"
            />
            <v-input
              v-if="movementMode === 'Straight'"
              type="number"
              :value="speed"
              label="Speed (mm/sec)"
              @input="speed = $event.detail.value"
            />
            <div
              v-if="movementMode === 'Straight'"
              :class="{ 'pointer-events-none opacity-50': movementType === 'Continuous' }"
            >
              <v-input
                type="number"
                :value="increment"
                :readonly="movementType === 'Continuous' ? 'true' : 'false'"
                label="Distance (mm)"
                @input="increment = $event.detail.value"
              />
            </div>
            <v-input
              v-if="movementMode === 'Spin'"
              type="number"
              :value="spinSpeed"
              label="Speed (deg/sec)"
              @input="spinSpeed = $event.detail.value"
            />
            <v-radio
              v-if="movementMode === 'Spin'"
              label="Movement Type"
              options="Clockwise, Counterclockwise"
              :selected="spinType"
              @input="setSpinType($event.detail.value)"
            />
            <div
              v-if="movementMode === 'Spin'"
              class="w-72 pl-6"
            >
              <v-slider
                :min="0"
                :max="360"
                :step="90"
                suffix="°"
                label="Angle"
                :value="angle"
                @input="angle = $event.detail.value"
              />
            </div>
          </div>
        </div>
        <div class="self-end">
          <v-button
            icon="play-circle-filled"
            variant="success"
            label="RUN"
            @click="baseRun()"
          />
        </div>
      </div>
    </div>
  </v-collapse>
</template>
