// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ttnv2

import (
	"context"
	"time"

	ttnsdk "github.com/TheThingsNetwork/go-app-sdk"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
)

// deviceManagerWithRetry is a ttnsdk.DeviceManager that retries Set() and Get() methods when a resource exhausted or service unavailable error is returned.
type deviceManagerWithRetry struct {
	ttnsdk.DeviceManager

	ctx        context.Context
	maxRetries uint
	backoff    time.Duration
}

func (d *deviceManagerWithRetry) shouldRetry(ctx context.Context, err error, attempt uint) (bool, time.Duration) {
	if err == nil || attempt >= d.maxRetries {
		return false, 0
	}
	if err, ok := errors.From(err); ok && (errors.IsResourceExhausted(err) || errors.IsUnavailable(err)) {
		penalty := d.backoff * time.Duration(attempt)
		log.FromContext(d.ctx).WithError(err).WithField("attempt", attempt).Debugf("Non-fatal error, will retry after %v", penalty)
		return true, penalty
	}
	return false, 0
}

func (d *deviceManagerWithRetry) get(ctx context.Context, devID string, attempt uint) (*ttnsdk.Device, error) {
	dev, err := d.DeviceManager.Get(devID)
	if retry, penalty := d.shouldRetry(ctx, err, attempt); retry {
		select {
		case <-time.After(penalty):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		return d.get(ctx, devID, attempt+1)
	}
	return dev, err
}

func (d *deviceManagerWithRetry) set(ctx context.Context, dev *ttnsdk.Device, attempt uint) error {
	err := d.DeviceManager.Set(dev)
	if retry, penalty := d.shouldRetry(ctx, err, attempt); retry {
		select {
		case <-time.After(penalty):
		case <-ctx.Done():
			return ctx.Err()
		}
		return d.set(ctx, dev, attempt+1)
	}
	return err
}

func (d *deviceManagerWithRetry) Get(devID string) (*ttnsdk.Device, error) {
	return d.get(log.NewContextWithFields(d.ctx, log.Fields("device_id", devID, "method", "Get")), devID, 1)
}

func (d *deviceManagerWithRetry) Set(dev *ttnsdk.Device) error {
	return d.set(log.NewContextWithFields(d.ctx, log.Fields("dev_eui", dev.DevEUI, "device_id", dev.DevID, "method", "Set")), dev, 1)
}

var _ ttnsdk.DeviceManager = &deviceManagerWithRetry{}
