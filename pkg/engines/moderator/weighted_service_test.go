package moderator

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTextModeratorService is a mock implementation of TextModeratorService
type MockTextModeratorService struct {
	mock.Mock
}

func (m *MockTextModeratorService) Allow(ctx context.Context, text []rune) error {
	args := m.Called(ctx, text)
	return args.Error(0)
}

func (m *MockTextModeratorService) MaxRuneLen() int {
	args := m.Called()
	return args.Int(0)
}

func TestNewWeightedModeratorService(t *testing.T) {
	tests := []struct {
		name          string
		services      []WeightedModeratorItem
		expectError   bool
		errorContains string
		validate      func(t *testing.T, service *WeightedModeratorService)
	}{
		{
			name:          "empty services list",
			services:      []WeightedModeratorItem{},
			expectError:   true,
			errorContains: "services must have at least one item",
		},
		{
			name: "negative weight",
			services: []WeightedModeratorItem{
				{
					Name:    "service1",
					Weight:  -1,
					Service: &MockTextModeratorService{},
				},
			},
			expectError:   true,
			errorContains: "weight must be >= 0",
		},
		{
			name: "single service with positive weight",
			services: []WeightedModeratorItem{
				{
					Name:   "service1",
					Weight: 100,
					Service: func() *MockTextModeratorService {
						mock := &MockTextModeratorService{}
						mock.On("MaxRuneLen").Return(10000)
						return mock
					}(),
				},
			},
			expectError: false,
			validate: func(t *testing.T, service *WeightedModeratorService) {
				assert.NotNil(t, service)
				assert.Equal(t, 10000, service.MaxRuneLen())
			},
		},
		{
			name: "multiple services with different weights",
			services: []WeightedModeratorItem{
				{
					Name:   "service1",
					Weight: 50,
					Service: func() *MockTextModeratorService {
						mock := &MockTextModeratorService{}
						mock.On("MaxRuneLen").Return(10000)
						return mock
					}(),
				},
				{
					Name:   "service2",
					Weight: 30,
					Service: func() *MockTextModeratorService {
						mock := &MockTextModeratorService{}
						mock.On("MaxRuneLen").Return(5000)
						return mock
					}(),
				},
				{
					Name:   "service3",
					Weight: 20,
					Service: func() *MockTextModeratorService {
						mock := &MockTextModeratorService{}
						mock.On("MaxRuneLen").Return(8000)
						return mock
					}(),
				},
			},
			expectError: false,
			validate: func(t *testing.T, service *WeightedModeratorService) {
				assert.NotNil(t, service)
				// MaxRuneLen should be the minimum of all service limits
				assert.Equal(t, 5000, service.MaxRuneLen())
			},
		},
		{
			name: "all weights are zero - should set default weight",
			services: []WeightedModeratorItem{
				{
					Name:   "service1",
					Weight: 0,
					Service: func() *MockTextModeratorService {
						mock := &MockTextModeratorService{}
						mock.On("MaxRuneLen").Return(10000)
						return mock
					}(),
				},
				{
					Name:   "service2",
					Weight: 0,
					Service: func() *MockTextModeratorService {
						mock := &MockTextModeratorService{}
						mock.On("MaxRuneLen").Return(5000)
						return mock
					}(),
				},
			},
			expectError: false,
			validate: func(t *testing.T, service *WeightedModeratorService) {
				assert.NotNil(t, service)
				assert.Equal(t, 5000, service.MaxRuneLen())
			},
		},
		{
			name: "mixed zero and non-zero weights",
			services: []WeightedModeratorItem{
				{
					Name:   "service1",
					Weight: 100,
					Service: func() *MockTextModeratorService {
						mock := &MockTextModeratorService{}
						mock.On("MaxRuneLen").Return(10000)
						return mock
					}(),
				},
				{
					Name:   "service2",
					Weight: 0,
					Service: func() *MockTextModeratorService {
						mock := &MockTextModeratorService{}
						mock.On("MaxRuneLen").Return(5000)
						return mock
					}(),
				},
			},
			expectError: false,
			validate: func(t *testing.T, service *WeightedModeratorService) {
				assert.NotNil(t, service)
				assert.Equal(t, 5000, service.MaxRuneLen())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewWeightedModeratorService(tt.services)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, service)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)
				if tt.validate != nil {
					tt.validate(t, service)
				}
			}
		})
	}
}

func TestWeightedModeratorService_Allow(t *testing.T) {
	t.Run("successful moderation", func(t *testing.T) {
		mockService1 := &MockTextModeratorService{}
		mockService1.On("MaxRuneLen").Return(10000)
		mockService1.On("Allow", mock.Anything, mock.Anything).Maybe().Return(nil)

		mockService2 := &MockTextModeratorService{}
		mockService2.On("MaxRuneLen").Return(10000)
		mockService2.On("Allow", mock.Anything, mock.Anything).Maybe().Return(nil)

		services := []WeightedModeratorItem{
			{Name: "service1", Weight: 50, Service: mockService1},
			{Name: "service2", Weight: 50, Service: mockService2},
		}

		weightedService, err := NewWeightedModeratorService(services)
		assert.NoError(t, err)
		assert.NotNil(t, weightedService)

		ctx := context.Background()
		text := []rune("test content")
		err = weightedService.Allow(ctx, text)
		assert.NoError(t, err)

		// Verify that at least one of the services was called
		// Weighted round robin selects one service per call, so we use Maybe() to make calls optional
		mockService1.AssertExpectations(t)
		mockService2.AssertExpectations(t)

		// Verify that exactly one service's Allow method was called
		allowCallCount1 := 0
		allowCallCount2 := 0
		for _, call := range mockService1.Calls {
			if call.Method == "Allow" {
				allowCallCount1++
			}
		}
		for _, call := range mockService2.Calls {
			if call.Method == "Allow" {
				allowCallCount2++
			}
		}
		assert.Equal(t, 1, allowCallCount1+allowCallCount2, "Exactly one service should be called")
	})

	t.Run("service returns error", func(t *testing.T) {
		expectedError := errors.New("content blocked")
		mockService := &MockTextModeratorService{}
		mockService.On("MaxRuneLen").Return(10000)
		mockService.On("Allow", mock.Anything, mock.Anything).Return(expectedError)

		services := []WeightedModeratorItem{
			{Name: "service1", Weight: 100, Service: mockService},
		}

		weightedService, err := NewWeightedModeratorService(services)
		assert.NoError(t, err)

		ctx := context.Background()
		text := []rune("test content")
		err = weightedService.Allow(ctx, text)
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)

		mockService.AssertExpectations(t)
	})
}

func TestWeightedModeratorService_MaxRuneLen(t *testing.T) {
	t.Run("returns minimum of all service limits", func(t *testing.T) {
		mockService1 := &MockTextModeratorService{}
		mockService1.On("MaxRuneLen").Return(20000)

		mockService2 := &MockTextModeratorService{}
		mockService2.On("MaxRuneLen").Return(10000)

		mockService3 := &MockTextModeratorService{}
		mockService3.On("MaxRuneLen").Return(15000)

		services := []WeightedModeratorItem{
			{Name: "service1", Weight: 33, Service: mockService1},
			{Name: "service2", Weight: 33, Service: mockService2},
			{Name: "service3", Weight: 34, Service: mockService3},
		}

		weightedService, err := NewWeightedModeratorService(services)
		assert.NoError(t, err)
		assert.Equal(t, 10000, weightedService.MaxRuneLen())
	})

	t.Run("single service", func(t *testing.T) {
		mockService := &MockTextModeratorService{}
		mockService.On("MaxRuneLen").Return(5000)

		services := []WeightedModeratorItem{
			{Name: "service1", Weight: 100, Service: mockService},
		}

		weightedService, err := NewWeightedModeratorService(services)
		assert.NoError(t, err)
		assert.Equal(t, 5000, weightedService.MaxRuneLen())
	})
}

func TestWeightedModeratorService_getNextService(t *testing.T) {
	t.Run("weighted round robin selection", func(t *testing.T) {
		mockService1 := &MockTextModeratorService{}
		mockService1.On("MaxRuneLen").Return(10000)

		mockService2 := &MockTextModeratorService{}
		mockService2.On("MaxRuneLen").Return(10000)

		mockService3 := &MockTextModeratorService{}
		mockService3.On("MaxRuneLen").Return(10000)

		services := []WeightedModeratorItem{
			{Name: "service1", Weight: 50, Service: mockService1},
			{Name: "service2", Weight: 30, Service: mockService2},
			{Name: "service3", Weight: 20, Service: mockService3},
		}

		weightedService, err := NewWeightedModeratorService(services)
		assert.NoError(t, err)

		// Call getNextService multiple times to test weighted round robin
		selectedServices := make(map[string]int)
		for i := 0; i < 100; i++ {
			name, _ := weightedService.getNextService()
			selectedServices[name]++
		}

		// Verify that all services were selected at least once
		assert.Greater(t, selectedServices["service1"], 0)
		assert.Greater(t, selectedServices["service2"], 0)
		assert.Greater(t, selectedServices["service3"], 0)

		// Verify that service1 (weight 50) is selected more often than service3 (weight 20)
		assert.Greater(t, selectedServices["service1"], selectedServices["service3"])
	})
}
