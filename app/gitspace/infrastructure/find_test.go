// Copyright 2023 Harness, Inc.
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

package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/harness/gitness/infraprovider"
	"github.com/harness/gitness/types"
	"github.com/harness/gitness/types/enum"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInfraProvisioner_Find(t *testing.T) {
	type fields struct {
		infraProviderConfigStore   *mockInfraProviderConfigStore
		infraProviderResourceStore *mockInfraProviderResourceStore
		providerFactory            *mockProviderFactory
		infraProviderTemplateStore *mockInfraProviderTemplateStore
		infraProvisionedStore      *mockInfraProvisionedStore
		config                     *Config
	}
	type args struct {
		ctx            context.Context
		gitspaceConfig types.GitspaceConfig
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *types.Infrastructure
		wantErr bool
	}{
		{
			name: "new provisioning type success",
			fields: fields{
				infraProviderConfigStore: func() *mockInfraProviderConfigStore {
					store := &mockInfraProviderConfigStore{}
					store.On("Find", mock.Anything, mock.Anything, mock.Anything).Return(&types.InfraProviderConfig{
						Type: enum.InfraProviderTypeDocker,
					}, nil)
					return store
				}(),
				providerFactory: func() *mockProviderFactory {
					factory := &mockProviderFactory{}
					provider := &mockInfraProvider{}
					provider.On("ProvisioningType").Return(enum.InfraProvisioningTypeNew)
					status := enum.InfraStatus("running")
					provider.On("FindInfraStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
						Return(&status, nil)
					factory.On("GetInfraProvider", mock.Anything).Return(provider, nil)
					return factory
				}(),
				infraProvisionedStore: func() *mockInfraProvisionedStore {
					store := &mockInfraProvisionedStore{}
					responseMetadata := `{"status": "running"}`
					store.On("FindLatestByGitspaceInstanceID", mock.Anything, mock.Anything).Return(&types.InfraProvisioned{
						ResponseMetadata: &responseMetadata,
					}, nil)
					return store
				}(),
				infraProviderTemplateStore: &mockInfraProviderTemplateStore{},
				config:                     &Config{},
			},
			args: args{
				ctx: context.Background(),
				gitspaceConfig: types.GitspaceConfig{
					GitspaceInstance: &types.GitspaceInstance{},
					IDE:              enum.IDETypeVSCode,
				},
			},
			want: &types.Infrastructure{
				Status:         "running",
				GitspaceScheme: "ssh",
			},
			wantErr: false,
		},
		{
			name: "existing provisioning type success",
			fields: fields{
				infraProviderConfigStore: func() *mockInfraProviderConfigStore {
					store := &mockInfraProviderConfigStore{}
					store.On("Find", mock.Anything, mock.Anything, mock.Anything).Return(&types.InfraProviderConfig{
						Type: enum.InfraProviderTypeDocker,
					}, nil)
					return store
				}(),
				providerFactory: func() *mockProviderFactory {
					factory := &mockProviderFactory{}
					provider := &mockInfraProvider{}
					provider.On("ProvisioningType").Return(enum.InfraProvisioningTypeExisting)
					provider.On("Find", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
						Return(&types.Infrastructure{Status: "running"}, nil)
					provider.On("TemplateParams").Return([]types.InfraProviderParameterSchema{})
					factory.On("GetInfraProvider", mock.Anything).Return(provider, nil)
					return factory
				}(),
				infraProvisionedStore:      &mockInfraProvisionedStore{},
				infraProviderTemplateStore: &mockInfraProviderTemplateStore{},
				config:                     &Config{},
			},
			args: args{
				ctx: context.Background(),
				gitspaceConfig: types.GitspaceConfig{
					GitspaceInstance: &types.GitspaceInstance{},
					IDE:              enum.IDETypeVSCode,
				},
			},
			want: &types.Infrastructure{
				Status:         "running",
				GitspaceScheme: "ssh",
			},
			wantErr: false,
		},
		{
			name: "error getting infra provider config",
			fields: fields{
				infraProviderConfigStore: func() *mockInfraProviderConfigStore {
					store := &mockInfraProviderConfigStore{}
					store.On("Find", mock.Anything, mock.Anything, mock.Anything).
						Return(nil, errors.New("some error"))
					return store
				}(),
				providerFactory:            &mockProviderFactory{},
				infraProvisionedStore:      &mockInfraProvisionedStore{},
				infraProviderTemplateStore: &mockInfraProviderTemplateStore{},
				config:                     &Config{},
			},
			args: args{
				ctx: context.Background(),
				gitspaceConfig: types.GitspaceConfig{
					GitspaceInstance: &types.GitspaceInstance{},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "existing provisioning type - error finding infra",
			fields: fields{
				infraProviderConfigStore: func() *mockInfraProviderConfigStore {
					store := &mockInfraProviderConfigStore{}
					store.On("Find", mock.Anything, mock.Anything, mock.Anything).Return(&types.InfraProviderConfig{
						Type: enum.InfraProviderTypeDocker,
					}, nil)
					return store
				}(),
				providerFactory: func() *mockProviderFactory {
					factory := &mockProviderFactory{}
					provider := &mockInfraProvider{}
					provider.On("ProvisioningType").Return(enum.InfraProvisioningTypeExisting)
					provider.On("Find", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
						Return(nil, errors.New("some error"))
					provider.On("TemplateParams").Return([]types.InfraProviderParameterSchema{})
					factory.On("GetInfraProvider", mock.Anything).Return(provider, nil)
					return factory
				}(),
				infraProvisionedStore:      &mockInfraProvisionedStore{},
				infraProviderTemplateStore: &mockInfraProviderTemplateStore{},
				config:                     &Config{},
			},
			args: args{
				ctx: context.Background(),
				gitspaceConfig: types.GitspaceConfig{
					GitspaceInstance: &types.GitspaceInstance{},
					IDE:              enum.IDETypeVSCode,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "new provisioning type - error finding infra status",
			fields: fields{
				infraProviderConfigStore: func() *mockInfraProviderConfigStore {
					store := &mockInfraProviderConfigStore{}
					store.On("Find", mock.Anything, mock.Anything, mock.Anything).Return(&types.InfraProviderConfig{
						Type: enum.InfraProviderTypeDocker,
					}, nil)
					return store
				}(),
				providerFactory: func() *mockProviderFactory {
					factory := &mockProviderFactory{}
					provider := &mockInfraProvider{}
					provider.On("ProvisioningType").Return(enum.InfraProvisioningTypeNew)
					provider.On("FindInfraStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
						Return(nil, errors.New("some error"))
					factory.On("GetInfraProvider", mock.Anything).Return(provider, nil)
					return factory
				}(),
				infraProvisionedStore: func() *mockInfraProvisionedStore {
					store := &mockInfraProvisionedStore{}
					responseMetadata := `{"status": "running"}`
					store.On("FindLatestByGitspaceInstanceID", mock.Anything, mock.Anything).Return(&types.InfraProvisioned{
						ResponseMetadata: &responseMetadata,
					}, nil)
					return store
				}(),
				infraProviderTemplateStore: &mockInfraProviderTemplateStore{},
				config:                     &Config{},
			},
			args: args{
				ctx: context.Background(),
				gitspaceConfig: types.GitspaceConfig{
					GitspaceInstance: &types.GitspaceInstance{},
					IDE:              enum.IDETypeVSCode,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error getting infra provider from factory",
			fields: fields{
				infraProviderConfigStore: func() *mockInfraProviderConfigStore {
					store := &mockInfraProviderConfigStore{}
					store.On("Find", mock.Anything, mock.Anything, mock.Anything).Return(&types.InfraProviderConfig{
						Type: enum.InfraProviderTypeDocker,
					}, nil)
					return store
				}(),
				providerFactory: func() *mockProviderFactory {
					factory := &mockProviderFactory{}
					factory.On("GetInfraProvider", mock.Anything).Return(nil, errors.New("some error"))
					return factory
				}(),
				infraProvisionedStore:      &mockInfraProvisionedStore{},
				infraProviderTemplateStore: &mockInfraProviderTemplateStore{},
				config:                     &Config{},
			},
			args: args{
				ctx: context.Background(),
				gitspaceConfig: types.GitspaceConfig{
					GitspaceInstance: &types.GitspaceInstance{},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "new provisioning type - error getting params",
			fields: fields{
				infraProviderConfigStore: func() *mockInfraProviderConfigStore {
					store := &mockInfraProviderConfigStore{}
					store.On("Find", mock.Anything, mock.Anything, mock.Anything).Return(&types.InfraProviderConfig{
						Type: enum.InfraProviderTypeDocker,
					}, nil)
					return store
				}(),
				providerFactory: func() *mockProviderFactory {
					factory := &mockProviderFactory{}
					provider := &mockInfraProvider{}
					provider.On("ProvisioningType").Return(enum.InfraProvisioningTypeNew)
					factory.On("GetInfraProvider", mock.Anything).Return(provider, nil)
					return factory
				}(),
				infraProvisionedStore: func() *mockInfraProvisionedStore {
					store := &mockInfraProvisionedStore{}
					store.On("FindLatestByGitspaceInstanceID", mock.Anything, mock.Anything).
						Return(nil, errors.New("some error"))
					return store
				}(),
				infraProviderTemplateStore: &mockInfraProviderTemplateStore{},
				config:                     &Config{},
			},
			args: args{
				ctx: context.Background(),
				gitspaceConfig: types.GitspaceConfig{
					GitspaceInstance: &types.GitspaceInstance{},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := InfraProvisioner{
				infraProviderConfigStore:   tt.fields.infraProviderConfigStore,
				infraProviderResourceStore: tt.fields.infraProviderResourceStore,
				providerFactory:            tt.fields.providerFactory,
				infraProviderTemplateStore: tt.fields.infraProviderTemplateStore,
				infraProvisionedStore:      tt.fields.infraProvisionedStore,
				config:                     tt.fields.config,
			}
			got, err := i.Find(tt.args.ctx, tt.args.gitspaceConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("Find() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_deserializeInfraProviderParams(t *testing.T) {
	type args struct {
		in string
	}
	tests := []struct {
		name    string
		args    args
		want    []types.InfraProviderParameter
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				in: `[{"name": "p1", "value": "v1"}]`,
			},
			want: []types.InfraProviderParameter{
				{Name: "p1", Value: "v1"},
			},
			wantErr: false,
		},
		{
			name: "invalid json",
			args: args{
				in: "invalid",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := deserializeInfraProviderParams(tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("deserializeInfraProviderParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_getGitspaceScheme(t *testing.T) {
	type args struct {
		ideType                  enum.IDEType
		gitspaceSchemeFromMetadata string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "vscode web",
			args: args{
				ideType:                  enum.IDETypeVSCodeWeb,
				gitspaceSchemeFromMetadata: "https",
			},
			want:    "https",
			wantErr: false,
		},
		{
			name: "vscode",
			args: args{
				ideType: enum.IDETypeVSCode,
			},
			want:    "ssh",
			wantErr: false,
		},
		{
			name: "unknown",
			args: args{
				ideType: "unknown",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getGitspaceScheme(tt.args.ideType, tt.args.gitspaceSchemeFromMetadata)
			if (err != nil) != tt.wantErr {
				t.Errorf("getGitspaceScheme() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

type mockInfraProviderConfigStore struct {
	mock.Mock
}

func (m *mockInfraProviderConfigStore) FindByType(ctx context.Context, spaceID int64, infraType enum.InfraProviderType, includeDeleted bool) (*types.InfraProviderConfig, error) {
	args := m.Called(ctx, spaceID, infraType, includeDeleted)
	return args.Get(0).(*types.InfraProviderConfig), args.Error(1)
}

func (m *mockInfraProviderConfigStore) Update(ctx context.Context, infraProviderConfig *types.InfraProviderConfig) error {
	args := m.Called(ctx, infraProviderConfig)
	return args.Error(0)
}

func (m *mockInfraProviderConfigStore) Find(ctx context.Context, id int64, includeDeleted bool) (*types.InfraProviderConfig, error) {
	args := m.Called(ctx, id, includeDeleted)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.InfraProviderConfig), args.Error(1)
}

func (m *mockInfraProviderConfigStore) List(ctx context.Context, filter *types.InfraProviderConfigFilter) ([]*types.InfraProviderConfig, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*types.InfraProviderConfig), args.Error(1)
}

func (m *mockInfraProviderConfigStore) FindByIdentifier(ctx context.Context, spaceID int64, identifier string) (*types.InfraProviderConfig, error) {
	args := m.Called(ctx, spaceID, identifier)
	return args.Get(0).(*types.InfraProviderConfig), args.Error(1)
}

func (m *mockInfraProviderConfigStore) Create(ctx context.Context, infraProviderConfig *types.InfraProviderConfig) error {
	args := m.Called(ctx, infraProviderConfig)
	return args.Error(0)
}

func (m *mockInfraProviderConfigStore) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type mockInfraProviderResourceStore struct {
	mock.Mock
}

func (m *mockInfraProviderResourceStore) List(ctx context.Context, infraProviderConfigID int64, filter types.ListQueryFilter) ([]*types.InfraProviderResource, error) {
	args := m.Called(ctx, infraProviderConfigID, filter)
	return args.Get(0).([]*types.InfraProviderResource), args.Error(1)
}

func (m *mockInfraProviderResourceStore) Find(ctx context.Context, id int64) (*types.InfraProviderResource, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*types.InfraProviderResource), args.Error(1)
}

func (m *mockInfraProviderResourceStore) FindByConfigAndIdentifier(ctx context.Context, spaceID int64, infraProviderConfigID int64, identifier string) (*types.InfraProviderResource, error) {
	args := m.Called(ctx, spaceID, infraProviderConfigID, identifier)
	return args.Get(0).(*types.InfraProviderResource), args.Error(1)
}

func (m *mockInfraProviderResourceStore) Create(ctx context.Context, infraProviderResource *types.InfraProviderResource) error {
	args := m.Called(ctx, infraProviderResource)
	return args.Error(0)
}

func (m *mockInfraProviderResourceStore) Update(ctx context.Context, infraProviderResource *types.InfraProviderResource) error {
	args := m.Called(ctx, infraProviderResource)
	return args.Error(0)
}

func (m *mockInfraProviderResourceStore) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type mockInfraProviderTemplateStore struct {
	mock.Mock
}

func (m *mockInfraProviderTemplateStore) FindByIdentifier(ctx context.Context, spaceID int64, identifier string) (*types.InfraProviderTemplate, error) {
	args := m.Called(ctx, spaceID, identifier)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.InfraProviderTemplate), args.Error(1)
}

func (m *mockInfraProviderTemplateStore) Find(ctx context.Context, id int64) (*types.InfraProviderTemplate, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*types.InfraProviderTemplate), args.Error(1)
}

func (m *mockInfraProviderTemplateStore) Create(ctx context.Context, infraProviderTemplate *types.InfraProviderTemplate) error {
	args := m.Called(ctx, infraProviderTemplate)
	return args.Error(0)
}

func (m *mockInfraProviderTemplateStore) Update(ctx context.Context, infraProviderTemplate *types.InfraProviderTemplate) error {
	args := m.Called(ctx, infraProviderTemplate)
	return args.Error(0)
}

func (m *mockInfraProviderTemplateStore) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type mockInfraProvisionedStore struct {
	mock.Mock
}

func (m *mockInfraProvisionedStore) Find(ctx context.Context, id int64) (*types.InfraProvisioned, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*types.InfraProvisioned), args.Error(1)
}

func (m *mockInfraProvisionedStore) FindAllLatestByGateway(ctx context.Context, gatewayHost string) ([]*types.InfraProvisionedGatewayView, error) {
	args := m.Called(ctx, gatewayHost)
	return args.Get(0).([]*types.InfraProvisionedGatewayView), args.Error(1)
}

func (m *mockInfraProvisionedStore) FindLatestByGitspaceInstanceID(ctx context.Context, gitspaceInstanceID int64) (*types.InfraProvisioned, error) {
	args := m.Called(ctx, gitspaceInstanceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.InfraProvisioned), args.Error(1)
}

func (m *mockInfraProvisionedStore) FindLatestByGitspaceInstanceIdentifier(ctx context.Context, spaceID int64, gitspaceInstanceIdentifier string) (*types.InfraProvisioned, error) {
	args := m.Called(ctx, spaceID, gitspaceInstanceIdentifier)
	return args.Get(0).(*types.InfraProvisioned), args.Error(1)
}

func (m *mockInfraProvisionedStore) FindStoppedInfraForGitspaceConfigIdentifier(ctx context.Context, gitspaceConfigIdentifier string) (*types.InfraProvisioned, error) {
	args := m.Called(ctx, gitspaceConfigIdentifier)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.InfraProvisioned), args.Error(1)
}

func (m *mockInfraProvisionedStore) Create(ctx context.Context, infraProvisioned *types.InfraProvisioned) error {
	args := m.Called(ctx, infraProvisioned)
	return args.Error(0)
}

func (m *mockInfraProvisionedStore) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockInfraProvisionedStore) Update(ctx context.Context, infraProvisioned *types.InfraProvisioned) error {
	args := m.Called(ctx, infraProvisioned)
	return args.Error(0)
}

type mockProviderFactory struct {
	mock.Mock
}

func (m *mockProviderFactory) GetInfraProvider(providerType enum.InfraProviderType) (infraprovider.InfraProvider, error) {
	args := m.Called(providerType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(infraprovider.InfraProvider), args.Error(1)
}

type mockInfraProvider struct {
	mock.Mock
}

func (m *mockInfraProvider) Provision(ctx context.Context, gitspaceConfig types.GitspaceConfig, agentPort int, requiredGitspacePorts []types.GitspacePort, inputParameters []types.InfraProviderParameter, configMetadata map[string]interface{}, existingInfrastructure types.Infrastructure) error {
	args := m.Called(ctx, gitspaceConfig, agentPort, requiredGitspacePorts, inputParameters, configMetadata, existingInfrastructure)
	return args.Error(0)
}

func (m *mockInfraProvider) Find(ctx context.Context, spaceID int64, spacePath string, gitspaceConfigIdentifier string, inputParameters []types.InfraProviderParameter) (*types.Infrastructure, error) {
	args := m.Called(ctx, spaceID, spacePath, gitspaceConfigIdentifier, inputParameters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Infrastructure), args.Error(1)
}

func (m *mockInfraProvider) FindInfraStatus(ctx context.Context, gitspaceConfigIdentifier string, gitspaceInstanceIdentifier string, inputParameters []types.InfraProviderParameter) (*enum.InfraStatus, error) {
	args := m.Called(ctx, gitspaceConfigIdentifier, gitspaceInstanceIdentifier, inputParameters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*enum.InfraStatus), args.Error(1)
}

func (m *mockInfraProvider) Stop(ctx context.Context, infra types.Infrastructure, gitspaceConfig types.GitspaceConfig, configMetadata map[string]interface{}) error {
	args := m.Called(ctx, infra, gitspaceConfig, configMetadata)
	return args.Error(0)
}

func (m *mockInfraProvider) CleanupInstanceResources(ctx context.Context, infra types.Infrastructure) error {
	args := m.Called(ctx, infra)
	return args.Error(0)
}

func (m *mockInfraProvider) Deprovision(ctx context.Context, infra types.Infrastructure, gitspaceConfig types.GitspaceConfig, canDeleteUserData bool, configMetadata map[string]interface{}, params []types.InfraProviderParameter) error {
	args := m.Called(ctx, infra, gitspaceConfig, canDeleteUserData, configMetadata, params)
	return args.Error(0)
}

func (m *mockInfraProvider) AvailableParams() []types.InfraProviderParameterSchema {
	args := m.Called()
	return args.Get(0).([]types.InfraProviderParameterSchema)
}

func (m *mockInfraProvider) UpdateParams(inputParameters []types.InfraProviderParameter, configMetaData map[string]interface{}) ([]types.InfraProviderParameter, error) {
	args := m.Called(inputParameters, configMetaData)
	return args.Get(0).([]types.InfraProviderParameter), args.Error(1)
}

func (m *mockInfraProvider) ValidateParams(inputParameters []types.InfraProviderParameter) error {
	args := m.Called(inputParameters)
	return args.Error(0)
}

func (m *mockInfraProvider) TemplateParams() []types.InfraProviderParameterSchema {
	args := m.Called()
	return args.Get(0).([]types.InfraProviderParameterSchema)
}

func (m *mockInfraProvider) ProvisioningType() enum.InfraProvisioningType {
	args := m.Called()
	return args.Get(0).(enum.InfraProvisioningType)
}

func (m *mockInfraProvider) UpdateConfig(infraProviderConfig *types.InfraProviderConfig) (*types.InfraProviderConfig, error) {
	args := m.Called(infraProviderConfig)
	return args.Get(0).(*types.InfraProviderConfig), args.Error(1)
}

func (m *mockInfraProvider) ValidateConfig(infraProviderConfig *types.InfraProviderConfig) error {
	args := m.Called(infraProviderConfig)
	return args.Error(0)
}

func (m *mockInfraProvider) GenerateSetupYAML(infraProviderConfig *types.InfraProviderConfig) (string, error) {
	args := m.Called(infraProviderConfig)
	return args.String(0), args.Error(1)
}

func TestInfraProvisioner_GetInfraFromStoredInfo(t *testing.T) {
	type fields struct {
		infraProvisionedStore *mockInfraProvisionedStore
	}
	type args struct {
		ctx            context.Context
		gitspaceConfig types.GitspaceConfig
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *types.Infrastructure
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				infraProvisionedStore: func() *mockInfraProvisionedStore {
					store := &mockInfraProvisionedStore{}
					responseMetadata := `{"status": "running"}`
					store.On("FindLatestByGitspaceInstanceID", mock.Anything, mock.Anything).Return(&types.InfraProvisioned{
						ResponseMetadata: &responseMetadata,
					}, nil)
					return store
				}(),
			},
			args: args{
				ctx: context.Background(),
				gitspaceConfig: types.GitspaceConfig{
					GitspaceInstance: &types.GitspaceInstance{ID: 1},
				},
			},
			want:    &types.Infrastructure{Status: "running"},
			wantErr: false,
		},
		{
			name: "find error",
			fields: fields{
				infraProvisionedStore: func() *mockInfraProvisionedStore {
					store := &mockInfraProvisionedStore{}
					store.On("FindLatestByGitspaceInstanceID", mock.Anything, mock.Anything).
						Return(nil, errors.New("some error"))
					return store
				}(),
			},
			args: args{
				ctx: context.Background(),
				gitspaceConfig: types.GitspaceConfig{
					GitspaceInstance: &types.GitspaceInstance{ID: 1},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "unmarshal error",
			fields: fields{
				infraProvisionedStore: func() *mockInfraProvisionedStore {
					store := &mockInfraProvisionedStore{}
					responseMetadata := `invalid`
					store.On("FindLatestByGitspaceInstanceID", mock.Anything, mock.Anything).Return(&types.InfraProvisioned{
						ResponseMetadata: &responseMetadata,
					}, nil)
					return store
				}(),
			},
			args: args{
				ctx: context.Background(),
				gitspaceConfig: types.GitspaceConfig{
					GitspaceInstance: &types.GitspaceInstance{ID: 1},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := InfraProvisioner{
				infraProvisionedStore: tt.fields.infraProvisionedStore,
			}
			got, err := i.GetInfraFromStoredInfo(tt.args.ctx, tt.args.gitspaceConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInfraFromStoredInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInfraProvisioner_GetStoppedInfraFromStoredInfo(t *testing.T) {
	validInfra := `{"key": "value"}`
	invalidInfra := "invalid-json"

	mockStore := &mockInfraProvisionedStore{}
	mockStore.On("FindStoppedInfraForGitspaceConfigIdentifier", mock.Anything, "valid").
		Return(&types.InfraProvisioned{ResponseMetadata: &validInfra}, nil)
	mockStore.On("FindStoppedInfraForGitspaceConfigIdentifier", mock.Anything, "invalid").
		Return(&types.InfraProvisioned{ResponseMetadata: &invalidInfra}, nil)
	mockStore.On("FindStoppedInfraForGitspaceConfigIdentifier", mock.Anything, "not-found").
		Return(nil, errors.New("not found"))

	i := InfraProvisioner{
		infraProvisionedStore: mockStore,
	}

	// Test case 1: Successfully find and unmarshal infra
	expectedInfra := types.Infrastructure{}
	json.Unmarshal([]byte(validInfra), &expectedInfra)
	infra, err := i.GetStoppedInfraFromStoredInfo(context.Background(), types.GitspaceConfig{Identifier: "valid"})
	assert.NoError(t, err)
	assert.Equal(t, expectedInfra, infra)

	// Test case 2: Fail to unmarshal infra
	infra, err = i.GetStoppedInfraFromStoredInfo(context.Background(), types.GitspaceConfig{Identifier: "invalid"})
	assert.Error(t, err)
	assert.Equal(t, types.Infrastructure{}, infra)

	// Test case 3: Fail to find infra
	infra, err = i.GetStoppedInfraFromStoredInfo(context.Background(), types.GitspaceConfig{Identifier: "not-found"})
	assert.Error(t, err)
	assert.Equal(t, types.Infrastructure{}, infra)
}
