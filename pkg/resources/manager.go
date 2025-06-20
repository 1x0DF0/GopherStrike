// Package resources provides resource management and cleanup for GopherStrike
package resources

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

// Resource represents a managed resource
type Resource interface {
	Close() error
	ID() string
}

// ResourceType represents the type of resource
type ResourceType string

const (
	FileResource       ResourceType = "file"
	NetworkResource    ResourceType = "network"
	HTTPClientResource ResourceType = "http_client"
	ConnectionResource ResourceType = "connection"
	ProcessResource    ResourceType = "process"
)

// Manager manages application resources and ensures cleanup
type Manager struct {
	mu        sync.RWMutex
	resources map[string]Resource
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	closed    bool
}

// NewManager creates a new resource manager
func NewManager(ctx context.Context) *Manager {
	managerCtx, cancel := context.WithCancel(ctx)
	
	m := &Manager{
		resources: make(map[string]Resource),
		ctx:       managerCtx,
		cancel:    cancel,
	}
	
	// Start cleanup goroutine
	m.wg.Add(1)
	go m.cleanupRoutine()
	
	return m
}

// cleanupRoutine periodically checks for resources to clean up
func (m *Manager) cleanupRoutine() {
	defer m.wg.Done()
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.cleanupStaleResources()
		}
	}
}

// cleanupStaleResources removes stale resources
func (m *Manager) cleanupStaleResources() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// In a real implementation, we would check for stale resources
	// For now, this is a placeholder
}

// Register registers a resource for management
func (m *Manager) Register(resource Resource) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.closed {
		return fmt.Errorf("manager is closed")
	}
	
	id := resource.ID()
	if _, exists := m.resources[id]; exists {
		return fmt.Errorf("resource with ID %s already registered", id)
	}
	
	m.resources[id] = resource
	return nil
}

// Unregister removes a resource from management
func (m *Manager) Unregister(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if resource, exists := m.resources[id]; exists {
		delete(m.resources, id)
		return resource.Close()
	}
	
	return fmt.Errorf("resource with ID %s not found", id)
}

// Get retrieves a registered resource
func (m *Manager) Get(id string) (Resource, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	resource, exists := m.resources[id]
	return resource, exists
}

// Close closes all managed resources
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.closed {
		return nil
	}
	
	m.closed = true
	m.cancel()
	
	var errors []error
	
	// Close all resources
	for id, resource := range m.resources {
		if err := resource.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close resource %s: %w", id, err))
		}
		delete(m.resources, id)
	}
	
	// Wait for cleanup routine to finish
	m.wg.Wait()
	
	if len(errors) > 0 {
		return fmt.Errorf("errors closing resources: %v", errors)
	}
	
	return nil
}

// FileResource represents a managed file
type FileResource struct {
	id   string
	file *os.File
}

// NewFileResource creates a new file resource
func NewFileResource(path string, flag int, perm os.FileMode) (*FileResource, error) {
	file, err := os.OpenFile(path, flag, perm)
	if err != nil {
		return nil, err
	}
	
	return &FileResource{
		id:   fmt.Sprintf("file:%s:%d", path, time.Now().UnixNano()),
		file: file,
	}, nil
}

// ID returns the resource ID
func (f *FileResource) ID() string {
	return f.id
}

// Close closes the file
func (f *FileResource) Close() error {
	if f.file != nil {
		return f.file.Close()
	}
	return nil
}

// File returns the underlying file
func (f *FileResource) File() *os.File {
	return f.file
}

// HTTPClientResource represents a managed HTTP client
type HTTPClientResource struct {
	id       string
	client   *http.Client
	transport *http.Transport
}

// NewHTTPClientResource creates a new HTTP client resource
func NewHTTPClientResource(timeout time.Duration) *HTTPClientResource {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
		DisableKeepAlives:   false,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}
	
	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
	
	return &HTTPClientResource{
		id:        fmt.Sprintf("http_client:%d", time.Now().UnixNano()),
		client:    client,
		transport: transport,
	}
}

// ID returns the resource ID
func (h *HTTPClientResource) ID() string {
	return h.id
}

// Close closes the HTTP client
func (h *HTTPClientResource) Close() error {
	if h.transport != nil {
		h.transport.CloseIdleConnections()
	}
	return nil
}

// Client returns the HTTP client
func (h *HTTPClientResource) Client() *http.Client {
	return h.client
}

// ConnectionResource represents a managed network connection
type ConnectionResource struct {
	id   string
	conn net.Conn
}

// NewConnectionResource creates a new connection resource
func NewConnectionResource(network, address string, timeout time.Duration) (*ConnectionResource, error) {
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}
	
	return &ConnectionResource{
		id:   fmt.Sprintf("conn:%s:%s:%d", network, address, time.Now().UnixNano()),
		conn: conn,
	}, nil
}

// ID returns the resource ID
func (c *ConnectionResource) ID() string {
	return c.id
}

// Close closes the connection
func (c *ConnectionResource) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Conn returns the network connection
func (c *ConnectionResource) Conn() net.Conn {
	return c.conn
}

// ResourcePool manages a pool of reusable resources
type ResourcePool struct {
	mu          sync.Mutex
	resources   chan Resource
	factory     func() (Resource, error)
	maxSize     int
	activeCount int
}

// NewResourcePool creates a new resource pool
func NewResourcePool(maxSize int, factory func() (Resource, error)) *ResourcePool {
	return &ResourcePool{
		resources: make(chan Resource, maxSize),
		factory:   factory,
		maxSize:   maxSize,
	}
}

// Get retrieves a resource from the pool
func (p *ResourcePool) Get() (Resource, error) {
	select {
	case resource := <-p.resources:
		return resource, nil
	default:
		p.mu.Lock()
		if p.activeCount >= p.maxSize {
			p.mu.Unlock()
			// Wait for a resource to become available
			resource := <-p.resources
			return resource, nil
		}
		p.activeCount++
		p.mu.Unlock()
		
		return p.factory()
	}
}

// Put returns a resource to the pool
func (p *ResourcePool) Put(resource Resource) {
	select {
	case p.resources <- resource:
		// Resource returned to pool
	default:
		// Pool is full, close the resource
		resource.Close()
		p.mu.Lock()
		p.activeCount--
		p.mu.Unlock()
	}
}

// Close closes all resources in the pool
func (p *ResourcePool) Close() error {
	close(p.resources)
	
	var errors []error
	for resource := range p.resources {
		if err := resource.Close(); err != nil {
			errors = append(errors, err)
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("errors closing pool resources: %v", errors)
	}
	
	return nil
}

// SafeClose safely closes any io.Closer
func SafeClose(closer io.Closer, name string) {
	if closer != nil {
		if err := closer.Close(); err != nil {
			// Log error but don't propagate
			fmt.Printf("Warning: failed to close %s: %v\n", name, err)
		}
	}
}

// WithResource executes a function with a managed resource
func WithResource(manager *Manager, resource Resource, fn func(Resource) error) error {
	if err := manager.Register(resource); err != nil {
		resource.Close()
		return err
	}
	
	defer manager.Unregister(resource.ID())
	
	return fn(resource)
}

// WithFile executes a function with a managed file
func WithFile(manager *Manager, path string, flag int, perm os.FileMode, fn func(*os.File) error) error {
	file, err := NewFileResource(path, flag, perm)
	if err != nil {
		return err
	}
	
	return WithResource(manager, file, func(r Resource) error {
		return fn(r.(*FileResource).File())
	})
}

// WithHTTPClient executes a function with a managed HTTP client
func WithHTTPClient(manager *Manager, timeout time.Duration, fn func(*http.Client) error) error {
	client := NewHTTPClientResource(timeout)
	
	return WithResource(manager, client, func(r Resource) error {
		return fn(r.(*HTTPClientResource).Client())
	})
}