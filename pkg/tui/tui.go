package tui

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"k8s-dashboard/pkg/k8s"

	"github.com/gdamore/tcell/v2"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// DataUpdate represents an update to resource data
type DataUpdate struct {
	ResourceType ResourceType
	Pods         []v1.Pod
	Deployments  []appsv1.Deployment
	Services     []v1.Service
	ConfigMaps   []v1.ConfigMap
	Namespaces   []v1.Namespace
	Error        error
}

// ResourceType represents different types of Kubernetes resources
type ResourceType int

const (
	ResourcePods ResourceType = iota
	ResourceDeployments
	ResourceServices
	ResourceConfigMaps
	ResourceNamespaces
)

// ViewMode represents different view modes
type ViewMode int

const (
	ViewModeList ViewMode = iota
	ViewModeDetails
	ViewModeYAML
	ViewModeLogs
	ViewModeRelationships
)

// LayoutMode represents different layout modes
type LayoutMode int

const (
	LayoutSingle LayoutMode = iota
	LayoutSplitVertical
	LayoutSplitHorizontal
)

// Theme represents a color theme
type Theme struct {
	background tcell.Color
	foreground tcell.Color
	header     tcell.Color
	accent     tcell.Color
	selected   tcell.Color
}

// Relationship represents a relationship between resources
type Relationship struct {
	From         string
	To           string
	RelationType string
}

// DefaultTheme returns the default color theme
func DefaultTheme() Theme {
	return Theme{
		background: tcell.ColorBlack,
		foreground: tcell.ColorWhite,
		header:     tcell.ColorBlue,
		accent:     tcell.ColorAqua,
		selected:   tcell.ColorYellow,
	}
}

// DarkTheme returns a dark color theme
func DarkTheme() Theme {
	return Theme{
		background: tcell.ColorBlack,
		foreground: tcell.ColorWhite,
		header:     tcell.ColorDarkBlue,
		accent:     tcell.ColorDarkCyan,
		selected:   tcell.ColorYellow,
	}
}

// LightTheme returns a light color theme
func LightTheme() Theme {
	return Theme{
		background: tcell.ColorWhite,
		foreground: tcell.ColorBlack,
		header:     tcell.ColorBlue,
		accent:     tcell.ColorDarkCyan,
		selected:   tcell.ColorRed,
	}
}

// SolarizedTheme returns a solarized color theme
func SolarizedTheme() Theme {
	return Theme{
		background: tcell.ColorBlack,
		foreground: tcell.ColorWhite,
		header:     0x073642, // base02
		accent:     0x2aa198, // cyan
		selected:   0xb58900, // yellow
	}
}

// DraculaTheme returns a dracula color theme
func DraculaTheme() Theme {
	return Theme{
		background: 0x282a36, // background
		foreground: 0xf8f8f2, // foreground
		header:     0x6272a4, // comment
		accent:     0x50fa7b, // green
		selected:   0xff79c6, // pink
	}
}

// NordTheme returns a nord-inspired color theme
func NordTheme() Theme {
	return Theme{
		background: 0x2e3440, // nord0
		foreground: 0xd8dee9, // nord4
		header:     0x5e81ac, // nord9
		accent:     0x88c0d0, // nord8
		selected:   0xebcb8b, // nord13
	}
}

// GruvboxTheme returns a gruvbox-inspired color theme
func GruvboxTheme() Theme {
	return Theme{
		background: 0x282828, // bg0
		foreground: 0xebdbb2, // fg
		header:     0x458588, // blue
		accent:     0x689d6a, // green
		selected:   0xd79921, // yellow
	}
}

// MonokaiTheme returns a monokai-inspired color theme
func MonokaiTheme() Theme {
	return Theme{
		background: 0x272822, // background
		foreground: 0xf8f8f2, // foreground
		header:     0x66d9ef, // blue
		accent:     0xa6e22e, // green
		selected:   0xfd971f, // orange
	}
}

// CyberpunkTheme returns a cyberpunk-inspired color theme
func CyberpunkTheme() Theme {
	return Theme{
		background: 0x0d0d0d, // dark background
		foreground: 0x00ff41, // matrix green
		header:     0xff0080, // magenta
		accent:     0x00ffff, // cyan
		selected:   0xffff00, // yellow
	}
}

// DisplayName returns the display name for a resource type
func (rt ResourceType) DisplayName() string {
	switch rt {
	case ResourcePods:
		return "Pods"
	case ResourceDeployments:
		return "Deployments"
	case ResourceServices:
		return "Services"
	case ResourceConfigMaps:
		return "ConfigMaps"
	case ResourceNamespaces:
		return "Namespaces"
	default:
		return "Unknown"
	}
}

// nextTheme cycles to the next available theme
func (t *TUI) nextTheme() {
	themes := []Theme{
		DefaultTheme(),
		DarkTheme(),
		LightTheme(),
		SolarizedTheme(),
		DraculaTheme(),
		NordTheme(),
		GruvboxTheme(),
		MonokaiTheme(),
		CyberpunkTheme(),
	}

	t.currentThemeIndex = (t.currentThemeIndex + 1) % len(themes)
	t.theme = themes[t.currentThemeIndex]

	// Force immediate redraw with clear
	t.screen.Clear()
	t.draw()
	t.screen.Sync()
	t.screen.Show()
}

// TUI represents the terminal user interface
type TUI struct {
	screen    tcell.Screen
	clientset kubernetes.Interface
	pods      []v1.Pod
	selected  int
	namespace string
	filter    string
	showHelp  bool
	loading   bool

	// Async loading
	loadingCounter int

	// Advanced filtering
	filterMode    bool
	columnFilters []string
	useRegex      bool
	caseSensitive bool
	inverseFilter bool

	// Theming
	currentThemeIndex int
	theme             Theme

	// Split-pane functionality
	splitRatio float64
	layoutMode LayoutMode

	// View modes
	currentView ResourceType
	viewMode    ViewMode

	// Resource data
	deployments []appsv1.Deployment
	services    []v1.Service
	configMaps  []v1.ConfigMap
	namespaces  []v1.Namespace

	// Scrolling
	detailsScroll       int
	logsScroll          int
	relationshipsScroll int

	// Relationships
	relationships []Relationship

	// Async data loading
	dataChan chan *DataUpdate
}

// NewTUI creates a new TUI instance
func NewTUI(clientset kubernetes.Interface) (*TUI, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, fmt.Errorf("failed to create screen: %v", err)
	}

	if err := screen.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize screen: %v", err)
	}

	screen.SetStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite))

	return &TUI{
		screen:    screen,
		clientset: clientset,
		selected:  0,
		namespace: "kube-system",
		filter:    "",
		showHelp:  false,
		loading:   false,

		// Async loading
		loadingCounter: 0,

		// Advanced filtering
		filterMode:    false,
		columnFilters: make([]string, 5), // 5 columns for pods
		useRegex:      false,
		caseSensitive: false,
		inverseFilter: false,

		// Theming
		currentThemeIndex: 0,
		theme:             DefaultTheme(),

		// Split-pane
		splitRatio: 0.5,
		layoutMode: LayoutSingle,

		// View modes
		currentView: ResourcePods,
		viewMode:    ViewModeList,

		// Resource data
		deployments: []appsv1.Deployment{},
		services:    []v1.Service{},
		configMaps:  []v1.ConfigMap{},

		// Scrolling
		detailsScroll:       0,
		logsScroll:          0,
		relationshipsScroll: 0,

		// Relationships
		relationships: []Relationship{},

		// Async data loading
		dataChan: make(chan *DataUpdate, 10),
	}, nil
}

// Run starts the TUI main loop
func (t *TUI) Run() error {
	defer t.screen.Fini()

	// Start data update handler
	go t.handleDataUpdates()

	// Initial data load
	if err := t.refreshData(); err != nil {
		return fmt.Errorf("failed to load data: %v", err)
	}

	// Main event loop
	for {
		t.draw()
		t.screen.Show()

		event := t.screen.PollEvent()
		switch ev := event.(type) {
		case *tcell.EventKey:
			if t.showHelp {
				// Any key exits help
				t.showHelp = false
				continue
			}

			// Handle view mode navigation
			if t.viewMode != ViewModeList {
				switch ev.Key() {
				case tcell.KeyEscape:
					t.viewMode = ViewModeList
					continue
				case tcell.KeyDown:
					switch t.viewMode {
					case ViewModeDetails, ViewModeYAML:
						t.detailsScroll++
					case ViewModeLogs:
						t.logsScroll++
					case ViewModeRelationships:
						t.relationshipsScroll++
					}
					continue
				case tcell.KeyUp:
					switch t.viewMode {
					case ViewModeDetails, ViewModeYAML:
						if t.detailsScroll > 0 {
							t.detailsScroll--
						}
					case ViewModeLogs:
						if t.logsScroll > 0 {
							t.logsScroll--
						}
					case ViewModeRelationships:
						if t.relationshipsScroll > 0 {
							t.relationshipsScroll--
						}
					}
					continue
				}
			}

			switch ev.Key() {
			case tcell.KeyEscape, tcell.KeyCtrlC:
				if t.viewMode != ViewModeList {
					t.viewMode = ViewModeList
				} else {
					return nil
				}
			case tcell.KeyDown, tcell.KeyRight:
				t.moveSelection(1)
			case tcell.KeyUp, tcell.KeyLeft:
				t.moveSelection(-1)
			case tcell.KeyEnter:
				if t.viewMode == ViewModeList {
					t.viewMode = ViewModeDetails
				}
			case tcell.KeyTab:
				t.currentView = ResourceType((int(t.currentView) + 1) % 5)
				t.selected = 0
			case tcell.KeyF5:
				t.refreshData()
			case tcell.KeyRune:
				switch ev.Rune() {
				case 'q':
					return nil
				case 'r':
					t.refreshData()
				case 'd':
					t.deleteSelectedResource()
				case 'n':
					t.changeNamespace()
				case 'c':
					t.createPodDialog()
				case 'h', '?':
					t.showHelp = true
				case '/':
					t.searchDialog()
				case 'f':
					t.clearFilter()
				case '1':
					t.currentView = ResourcePods
					t.selected = 0
				case '2':
					t.currentView = ResourceDeployments
					t.selected = 0
				case '3':
					t.currentView = ResourceServices
					t.selected = 0
				case '4':
					t.currentView = ResourceConfigMaps
					t.selected = 0
				case '5':
					t.currentView = ResourceNamespaces
					t.selected = 0
				case 'v':
					t.nextViewMode()
				case 'y':
					if t.viewMode == ViewModeDetails {
						t.viewMode = ViewModeYAML
					}
				case 'j':
					if t.viewMode == ViewModeDetails && t.currentView == ResourcePods {
						t.viewMode = ViewModeLogs
					}
				case 's':
					t.toggleSplitView()
				case 'S':
					t.switchSplitLayout()
				case 't', 'T':
					t.nextTheme()
				}
			}
		case *tcell.EventResize:
			t.screen.Sync()
		}
	}
}

// loadPods fetches pods from the current namespace
func (t *TUI) loadPods() error {
	pods, err := k8s.ListPods(t.clientset, t.namespace)
	if err != nil {
		klog.Errorf("Failed to list pods: %v", err)
		return err
	}

	t.pods = pods
	if t.selected >= len(t.pods) {
		t.selected = len(t.pods) - 1
	}
	if t.selected < 0 {
		t.selected = 0
	}

	return nil
}

// refreshData loads all resource types asynchronously
func (t *TUI) refreshData() error {
	t.loading = true
	t.loadingCounter = 5 // 5 resource types
	t.draw()
	t.screen.Show()

	// Clear existing data
	t.pods = nil
	t.deployments = nil
	t.services = nil
	t.configMaps = nil
	t.namespaces = nil

	// Start async loading
	go t.loadPodsAsync()
	go t.loadDeploymentsAsync()
	go t.loadServicesAsync()
	go t.loadConfigMapsAsync()
	go t.loadNamespacesAsync()

	return nil
}

// loadPodsAsync loads pods asynchronously
func (t *TUI) loadPodsAsync() {
	pods, err := k8s.ListPods(t.clientset, t.namespace)
	update := &DataUpdate{
		ResourceType: ResourcePods,
		Pods:         pods,
		Error:        err,
	}
	t.dataChan <- update
}

// loadDeploymentsAsync loads deployments asynchronously
func (t *TUI) loadDeploymentsAsync() {
	deployments, err := k8s.ListDeployments(t.clientset, t.namespace)
	update := &DataUpdate{
		ResourceType: ResourceDeployments,
		Deployments:  deployments,
		Error:        err,
	}
	t.dataChan <- update
}

// loadServicesAsync loads services asynchronously
func (t *TUI) loadServicesAsync() {
	services, err := k8s.ListServices(t.clientset, t.namespace)
	update := &DataUpdate{
		ResourceType: ResourceServices,
		Services:     services,
		Error:        err,
	}
	t.dataChan <- update
}

// loadConfigMapsAsync loads configmaps asynchronously
func (t *TUI) loadConfigMapsAsync() {
	configMaps, err := k8s.ListConfigMaps(t.clientset, t.namespace)
	update := &DataUpdate{
		ResourceType: ResourceConfigMaps,
		ConfigMaps:   configMaps,
		Error:        err,
	}
	t.dataChan <- update
}

// loadNamespacesAsync loads namespaces asynchronously
func (t *TUI) loadNamespacesAsync() {
	namespaces, err := k8s.ListNamespaces(t.clientset)
	update := &DataUpdate{
		ResourceType: ResourceNamespaces,
		Namespaces:   namespaces,
		Error:        err,
	}
	t.dataChan <- update
}

// loadDeployments fetches deployments from the current namespace
func (t *TUI) loadDeployments() error {
	deployments, err := k8s.ListDeployments(t.clientset, t.namespace)
	if err != nil {
		klog.Errorf("Failed to list deployments: %v", err)
		return err
	}
	t.deployments = deployments
	return nil
}

// loadServices fetches services from the current namespace
func (t *TUI) loadServices() error {
	services, err := k8s.ListServices(t.clientset, t.namespace)
	if err != nil {
		klog.Errorf("Failed to list services: %v", err)
		return err
	}
	t.services = services
	return nil
}

// loadConfigMaps fetches configmaps from the current namespace
func (t *TUI) loadConfigMaps() error {
	configMaps, err := k8s.ListConfigMaps(t.clientset, t.namespace)
	if err != nil {
		klog.Errorf("Failed to list configmaps: %v", err)
		return err
	}
	t.configMaps = configMaps
	return nil
}

// handleDataUpdates runs in a goroutine to process async data updates
func (t *TUI) handleDataUpdates() {
	for update := range t.dataChan {
		t.handleDataUpdate(update)
		// Trigger a redraw by sending a custom event or just continuing
		// For now, we'll rely on the main loop redrawing on events
	}
}

// handleDataUpdate processes a data update from async loading
func (t *TUI) handleDataUpdate(update *DataUpdate) {
	if update.Error != nil {
		klog.Errorf("Failed to load %v: %v", update.ResourceType, update.Error)
		// Could show error in UI
	}

	switch update.ResourceType {
	case ResourcePods:
		t.pods = update.Pods
		klog.Infof("Loaded %d pods", len(t.pods))
	case ResourceDeployments:
		t.deployments = update.Deployments
		klog.Infof("Loaded %d deployments", len(t.deployments))
	case ResourceServices:
		t.services = update.Services
		klog.Infof("Loaded %d services", len(t.services))
	case ResourceConfigMaps:
		t.configMaps = update.ConfigMaps
		klog.Infof("Loaded %d configmaps", len(t.configMaps))
	case ResourceNamespaces:
		t.namespaces = update.Namespaces
		klog.Infof("Loaded %d namespaces", len(t.namespaces))
	}

	// Decrement counter
	t.loadingCounter--
	if t.loadingCounter <= 0 {
		t.loading = false
		// Adjust selection if needed
		t.adjustSelection()
		klog.Infof("All resources loaded - Pods: %d, Deployments: %d, Services: %d, ConfigMaps: %d, Namespaces: %d in namespace: %s",
			len(t.pods), len(t.deployments), len(t.services), len(t.configMaps), len(t.namespaces), t.namespace)
	}
}

// adjustSelection ensures selected index is valid after data updates
func (t *TUI) adjustSelection() {
	var maxItems int
	switch t.currentView {
	case ResourcePods:
		maxItems = len(t.pods)
	case ResourceDeployments:
		maxItems = len(t.deployments)
	case ResourceServices:
		maxItems = len(t.services)
	case ResourceConfigMaps:
		maxItems = len(t.configMaps)
	}

	if t.selected >= maxItems {
		t.selected = maxItems - 1
	}
	if t.selected < 0 {
		t.selected = 0
	}
}

// moveSelection moves the selection up or down
func (t *TUI) moveSelection(delta int) {
	filtered := t.getFilteredResources()
	if len(filtered) == 0 {
		return
	}

	t.selected += delta
	if t.selected < 0 {
		t.selected = 0
	}
	if t.selected >= len(filtered) {
		t.selected = len(filtered) - 1
	}
}

// draw renders the TUI
func (t *TUI) draw() {
	t.screen.Clear()

	width, height := t.screen.Size()

	if t.showHelp {
		t.drawHelpScreen(width, height)
		return
	}

	if t.loading {
		t.drawLoadingScreen(width, height)
		return
	}

	// Handle different layout modes
	switch t.layoutMode {
	case LayoutSingle:
		t.drawSingleView(width, height)
	case LayoutSplitVertical:
		t.drawSplitVertical(width, height)
	case LayoutSplitHorizontal:
		t.drawSplitHorizontal(width, height)
	}
}

// drawSingleView draws the single-pane view
func (t *TUI) drawSingleView(width, height int) {
	// Handle different view modes
	switch t.viewMode {
	case ViewModeList:
		t.drawListView(width, height)
	case ViewModeDetails:
		t.drawDetailsView(width, height)
	case ViewModeYAML:
		t.drawYAMLView(width, height)
	case ViewModeLogs:
		t.drawLogsView(width, height)
	case ViewModeRelationships:
		t.drawRelationshipsView(width, height)
	}
}

// nextViewMode cycles through view modes
func (t *TUI) nextViewMode() {
	switch t.viewMode {
	case ViewModeList:
		t.viewMode = ViewModeDetails
	case ViewModeDetails:
		t.viewMode = ViewModeYAML
	case ViewModeYAML:
		if t.currentView == ResourcePods {
			t.viewMode = ViewModeLogs
		} else {
			t.viewMode = ViewModeList
		}
	case ViewModeLogs:
		t.viewMode = ViewModeRelationships
	case ViewModeRelationships:
		t.viewMode = ViewModeList
	}
}

// toggleSplitView toggles between single and split view modes
func (t *TUI) toggleSplitView() {
	if t.layoutMode == LayoutSingle {
		t.layoutMode = LayoutSplitVertical
	} else {
		t.layoutMode = LayoutSingle
	}
}

// switchSplitLayout switches between vertical and horizontal split
func (t *TUI) switchSplitLayout() {
	if t.layoutMode == LayoutSplitVertical {
		t.layoutMode = LayoutSplitHorizontal
	} else if t.layoutMode == LayoutSplitHorizontal {
		t.layoutMode = LayoutSplitVertical
	}
}

// deleteSelectedResource deletes the currently selected resource
func (t *TUI) deleteSelectedResource() {
	resource := t.getSelectedResource()
	if resource == nil {
		return
	}

	var name, resourceType string
	switch r := resource.(type) {
	case v1.Pod:
		name = r.Name
		resourceType = "pod"
	case appsv1.Deployment:
		name = r.Name
		resourceType = "deployment"
	case v1.Service:
		name = r.Name
		resourceType = "service"
	case v1.ConfigMap:
		name = r.Name
		resourceType = "configmap"
	default:
		return
	}

	// Show confirmation
	confirmMsg := fmt.Sprintf("Delete %s '%s'? (y/N)", resourceType, name)
	t.drawText(0, 1, 50, confirmMsg, tcell.StyleDefault.Background(tcell.ColorYellow).Foreground(tcell.ColorBlack))
	t.screen.Show()

	// Wait for confirmation
	event := t.screen.PollEvent()
	if ev, ok := event.(*tcell.EventKey); ok && ev.Rune() == 'y' {
		var err error
		switch r := resource.(type) {
		case v1.Pod:
			err = k8s.DeletePod(t.clientset, t.namespace, r.Name)
		case appsv1.Deployment:
			err = k8s.DeleteDeployment(t.clientset, t.namespace, r.Name)
		case v1.Service:
			err = k8s.DeleteService(t.clientset, t.namespace, r.Name)
		case v1.ConfigMap:
			err = k8s.DeleteConfigMap(t.clientset, t.namespace, r.Name)
		}

		if err != nil {
			klog.Errorf("Failed to delete %s: %v", resourceType, err)
			errorMsg := fmt.Sprintf("Error deleting %s: %v", resourceType, err)
			t.drawText(0, 3, 80, errorMsg, tcell.StyleDefault.Background(tcell.ColorRed).Foreground(tcell.ColorWhite))
			t.screen.Show()
			time.Sleep(2 * time.Second)
		} else {
			// Reload resources
			t.refreshData()
		}
	}
}

// drawSplitVertical draws a vertical split layout (left: list, right: details)
func (t *TUI) drawSplitVertical(width, height int) {
	leftWidth := int(float64(width) * t.splitRatio)
	rightWidth := width - leftWidth - 1 // -1 for separator

	// Draw left panel (list view)
	t.drawPanel(0, 0, leftWidth, height, true, ViewModeList)

	// Draw separator
	for y := 0; y < height; y++ {
		t.screen.SetContent(leftWidth, y, '│', nil, tcell.StyleDefault)
	}

	// Draw right panel (details view)
	t.drawPanel(leftWidth+1, 0, rightWidth, height, false, ViewModeDetails)
}

// drawSplitHorizontal draws a horizontal split layout (top: list, bottom: details)
func (t *TUI) drawSplitHorizontal(width, height int) {
	topHeight := int(float64(height) * t.splitRatio)
	bottomHeight := height - topHeight - 1 // -1 for separator

	// Draw top panel (list view)
	t.drawPanel(0, 0, width, topHeight, true, ViewModeList)

	// Draw separator
	for x := 0; x < width; x++ {
		t.screen.SetContent(x, topHeight, '─', nil, tcell.StyleDefault)
	}

	// Draw bottom panel (details view)
	t.drawPanel(0, topHeight+1, width, bottomHeight, false, ViewModeDetails)
}

// drawPanel draws a panel with specific view mode
func (t *TUI) drawPanel(x, y, width, height int, isLeftPanel bool, viewMode ViewMode) {
	// For now, use simplified implementation
	// In a full implementation, we'd need to offset all drawing operations

	switch viewMode {
	case ViewModeList:
		// Draw a simple list in the panel
		t.drawText(x, y, width, "List View", tcell.StyleDefault.Background(t.theme.header).Foreground(tcell.ColorWhite))
		filtered := t.getFilteredResources()
		for i := 0; i < len(filtered) && i < height-2; i++ {
			resource := filtered[i]
			name := t.getResourceName(resource)
			style := tcell.StyleDefault
			if isLeftPanel && i == t.selected {
				style = style.Background(t.theme.selected)
			}
			t.drawText(x, y+1+i, width, name, style)
		}
	case ViewModeDetails:
		// Draw details of selected resource
		resource := t.getSelectedResource()
		if resource != nil {
			t.drawText(x, y, width, "Details View", tcell.StyleDefault.Background(t.theme.header).Foreground(tcell.ColorWhite))
			details := t.getResourceDetails(resource)
			for i, line := range details {
				if i >= height-2 {
					break
				}
				t.drawText(x, y+1+i, width, line, tcell.StyleDefault)
			}
		}
	}
}

// drawListView draws the resource list view
func (t *TUI) drawListView(width, height int) {
	// Draw header (now 5 lines tall)
	t.drawHeader(width)

	// Draw search bar if filter is active
	if t.filter != "" || t.filterMode {
		t.drawSearchBar(width, 5)
	}

	// Draw main content area
	contentStartY := 6
	if t.filter != "" || t.filterMode {
		contentStartY = 8
	}
	contentHeight := height - contentStartY - 2 // Leave space for status and footer

	t.drawResourceTable(width, contentHeight, contentStartY)

	// Draw status bar
	t.drawStatusBar(width, height-2)

	// Draw footer
	t.drawFooter(width, height-1)
}

// drawHeader draws the header with resource tabs
func (t *TUI) drawHeader(width int) {
	// Draw top border
	topBorder := "┌" + strings.Repeat("─", width-2) + "┐"
	t.drawText(0, 0, width, topBorder, tcell.StyleDefault.Foreground(t.theme.accent))

	// Title with better styling
	title := " 🚀 KGO - Kubernetes Dashboard "
	padding := (width - len(title)) / 2
	if padding < 0 {
		padding = 0
	}
	titleBar := "│" + strings.Repeat(" ", padding) + title + strings.Repeat(" ", width-padding-len(title)-1) + "│"
	headerStyle := tcell.StyleDefault.Background(t.theme.header).Foreground(tcell.ColorWhite).Bold(true)
	t.drawText(0, 1, width, titleBar, headerStyle)

	// Separator line
	sepLine := "├" + strings.Repeat("─", width-2) + "┤"
	t.drawText(0, 2, width, sepLine, tcell.StyleDefault.Foreground(t.theme.accent))

	// Resource tabs with better styling
	tabs := []string{" 1.Pods ", " 2.Deployments ", " 3.Services ", " 4.ConfigMaps ", " 5.Namespaces "}
	tabsY := 3

	x := 0
	for i, tab := range tabs {
		var style tcell.Style
		if ResourceType(i) == t.currentView {
			// Active tab styling
			style = tcell.StyleDefault.Background(t.theme.selected).Foreground(tcell.ColorBlack).Bold(true)
			// Add brackets around active tab
			tab = "▶" + tab[1:] + "◀"
		} else {
			// Inactive tab styling
			style = tcell.StyleDefault.Background(t.theme.background).Foreground(t.theme.foreground)
		}
		t.drawText(x, tabsY, len(tab), tab, style)
		x += len(tab)
	}

	// Bottom border for header section
	bottomBorder := "├" + strings.Repeat("─", width-2) + "┤"
	t.drawText(0, 4, width, bottomBorder, tcell.StyleDefault.Foreground(t.theme.accent))
}

// drawResourceTable draws the resource table for current view
func (t *TUI) drawResourceTable(width, height, startY int) {
	filtered := t.getFilteredResources()

	if len(filtered) == 0 {
		t.drawText(0, startY, width, "No resources found", tcell.StyleDefault)
		return
	}

	// Get table headers and column widths based on resource type
	headers := t.getTableHeaders()
	colWidths := t.getColumnWidths(width, len(headers))

	// Draw table header with enhanced styling
	headerY := startY
	headerText := "┌" + strings.Repeat("─", width-2) + "┐"
	t.drawText(0, headerY, width, headerText, tcell.StyleDefault.Foreground(t.theme.accent))

	headerLine := "│ "
	for i, header := range headers {
		headerLine += fmt.Sprintf("%-*s", colWidths[i], header)
		if i < len(headers)-1 {
			headerLine += " │ "
		}
	}
	headerLine += " │"
	t.drawText(0, headerY+1, width, headerLine, tcell.StyleDefault.Background(t.theme.header).Foreground(tcell.ColorWhite).Bold(true))

	// Draw separator with enhanced styling
	sepLine := "├" + strings.Repeat("─", width-2) + "┤"
	t.drawText(0, headerY+2, width, sepLine, tcell.StyleDefault.Foreground(t.theme.accent))

	// Draw resources with alternating row colors
	resourceStartY := headerY + 3
	for i, resource := range filtered {
		if i >= height-5 { // Leave space for borders and footer
			break
		}

		y := resourceStartY + i
		style := tcell.StyleDefault

		// Highlight selected resource
		if i == t.selected {
			style = style.Background(t.theme.selected).Foreground(tcell.ColorBlack).Bold(true)
		} else {
			// Alternating row colors for better readability
			if i%2 == 0 {
				style = style.Background(t.theme.background)
			} else {
				// Slightly different background for alternate rows
				style = style.Background(tcell.ColorBlack)
			}
			style = style.Foreground(t.theme.foreground)
		}

		line := t.formatResourceLine(resource, colWidths)
		t.drawText(0, y, width, line, style)
	}

	// Draw bottom border
	if len(filtered) > 0 && len(filtered) < height-5 {
		bottomY := resourceStartY + len(filtered)
		bottomBorder := "└" + strings.Repeat("─", width-2) + "┘"
		t.drawText(0, bottomY, width, bottomBorder, tcell.StyleDefault.Foreground(t.theme.accent))
	}
}

// drawTitleBar draws the application title
func (t *TUI) drawTitleBar(width int) {
	title := " 🚀 Kubernetes Dashboard "
	padding := (width - len(title)) / 2
	if padding < 0 {
		padding = 0
	}

	titleBar := strings.Repeat("═", padding) + title + strings.Repeat("═", width-padding-len(title))
	t.drawText(0, 0, width, titleBar, tcell.StyleDefault.Background(tcell.ColorDarkBlue).Foreground(tcell.ColorWhite).Bold(true))
}

// drawSearchBar draws the search/filter bar
func (t *TUI) drawSearchBar(width, y int) {
	searchText := fmt.Sprintf(" 🔍 Filter: %s ", t.filter)
	if len(searchText) < width {
		searchText += strings.Repeat(" ", width-len(searchText))
	}
	t.drawText(0, y, width, searchText, tcell.StyleDefault.Background(tcell.ColorDarkCyan).Foreground(tcell.ColorWhite))
}

// drawPodTable draws the pod listing table
func (t *TUI) drawPodTable(width, height, startY int) {
	// Table headers with borders
	headers := []string{"Name", "Status", "Ready", "Age", "Node"}
	colWidths := []int{24, 11, 7, 11, 15}

	// Draw table header
	headerY := startY
	headerText := "┌" + strings.Repeat("─", width-2) + "┐"
	t.drawText(0, headerY, width, headerText, tcell.StyleDefault.Foreground(tcell.ColorGray))

	headerLine := "│ "
	for i, header := range headers {
		headerLine += fmt.Sprintf("%-*s", colWidths[i], header)
		if i < len(headers)-1 {
			headerLine += " │ "
		}
	}
	headerLine += " │"
	t.drawText(0, headerY+1, width, headerLine, tcell.StyleDefault.Background(tcell.ColorGray).Foreground(tcell.ColorBlack).Bold(true))

	// Draw separator
	sepLine := "├" + strings.Repeat("─", width-2) + "┤"
	t.drawText(0, headerY+2, width, sepLine, tcell.StyleDefault.Foreground(tcell.ColorGray))

	// Get filtered pods
	filteredPods := t.getFilteredPods()

	// Draw pods
	podStartY := headerY + 3
	for i, pod := range filteredPods {
		if i >= height-4 { // Leave space for borders
			break
		}

		y := podStartY + i
		style := tcell.StyleDefault

		// Highlight selected pod
		if i == t.selected && len(filteredPods) > 0 {
			style = style.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack).Bold(true)
		}

		// Color code status
		statusStyle := t.getPodStatusStyle(pod.Status.Phase)
		if i == t.selected {
			statusStyle = style
		}

		line := t.formatPodTableLine(pod, colWidths)
		t.drawText(0, y, width, line, statusStyle)
	}

	// Draw table bottom border
	if len(filteredPods) < height-4 {
		bottomY := podStartY + len(filteredPods)
		bottomLine := "└" + strings.Repeat("─", width-2) + "┘"
		t.drawText(0, bottomY, width, bottomLine, tcell.StyleDefault.Foreground(tcell.ColorGray))
	}
}

// getFilteredPods returns pods filtered by the current filter
func (t *TUI) getFilteredPods() []v1.Pod {
	if t.filter == "" {
		return t.pods
	}

	var filtered []v1.Pod
	for _, pod := range t.pods {
		if strings.Contains(strings.ToLower(pod.Name), strings.ToLower(t.filter)) {
			filtered = append(filtered, pod)
		}
	}
	return filtered
}

// getPodStatusStyle returns appropriate style for pod status
func (t *TUI) getPodStatusStyle(phase v1.PodPhase) tcell.Style {
	switch phase {
	case v1.PodRunning:
		return tcell.StyleDefault.Foreground(tcell.ColorGreen)
	case v1.PodPending:
		return tcell.StyleDefault.Foreground(tcell.ColorYellow)
	case v1.PodFailed:
		return tcell.StyleDefault.Foreground(tcell.ColorRed)
	case v1.PodSucceeded:
		return tcell.StyleDefault.Foreground(tcell.ColorBlue)
	default:
		return tcell.StyleDefault.Foreground(tcell.ColorGray)
	}
}

// formatPodTableLine formats a pod into a bordered table line
func (t *TUI) formatPodTableLine(pod v1.Pod, colWidths []int) string {
	name := pod.Name
	if len(name) > colWidths[0] {
		name = name[:colWidths[0]-3] + "..."
	}

	status := string(pod.Status.Phase)
	status = fmt.Sprintf("%-*s", colWidths[1], status)

	ready := t.getReadyCount(pod)
	ready = fmt.Sprintf("%-*s", colWidths[2], ready)

	age := t.formatDuration(time.Since(pod.CreationTimestamp.Time))
	age = fmt.Sprintf("%-*s", colWidths[3], age)

	node := pod.Spec.NodeName
	if node == "" {
		node = "<none>"
	}
	if len(node) > colWidths[4] {
		node = node[:colWidths[4]-3] + "..."
	}

	return fmt.Sprintf("│ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │",
		colWidths[0], name,
		colWidths[1], status,
		colWidths[2], ready,
		colWidths[3], age,
		colWidths[4], node)
}

// getFilteredResources returns filtered resources based on current view and filters
func (t *TUI) getFilteredResources() []interface{} {
	var resources []interface{}

	// Get resources based on current view
	switch t.currentView {
	case ResourcePods:
		for _, pod := range t.pods {
			resources = append(resources, pod)
		}
	case ResourceDeployments:
		for _, dep := range t.deployments {
			resources = append(resources, dep)
		}
	case ResourceServices:
		for _, svc := range t.services {
			resources = append(resources, svc)
		}
	case ResourceConfigMaps:
		for _, cm := range t.configMaps {
			resources = append(resources, cm)
		}
	case ResourceNamespaces:
		for _, ns := range t.namespaces {
			resources = append(resources, ns)
		}
	}

	// Apply filters
	if t.filter == "" && !t.filterMode {
		return resources
	}

	var filtered []interface{}
	for _, resource := range resources {
		if t.matchesFilter(resource) {
			filtered = append(filtered, resource)
		}
	}

	return filtered
}

// matchesFilter checks if a resource matches the current filter
func (t *TUI) matchesFilter(resource interface{}) bool {
	if t.filter == "" && !t.filterMode {
		return true
	}

	name := t.getResourceName(resource)
	if name == "" {
		return false
	}

	// Simple filter mode
	if !t.filterMode {
		if t.caseSensitive {
			return strings.Contains(name, t.filter)
		}
		return strings.Contains(strings.ToLower(name), strings.ToLower(t.filter))
	}

	// Advanced filter mode - check global filter and column filters
	match := true

	// Global filter
	if t.filter != "" {
		if t.useRegex {
			// TODO: Implement regex matching
			if t.caseSensitive {
				match = strings.Contains(name, t.filter)
			} else {
				match = strings.Contains(strings.ToLower(name), strings.ToLower(t.filter))
			}
		} else {
			if t.caseSensitive {
				match = strings.Contains(name, t.filter)
			} else {
				match = strings.Contains(strings.ToLower(name), strings.ToLower(t.filter))
			}
		}
	}

	// Column-specific filters
	headers := t.getTableHeaders()
	for i, colFilter := range t.columnFilters {
		if colFilter == "" || i >= len(headers) {
			continue
		}

		colValue := t.getResourceColumnValue(resource, i)
		if t.useRegex {
			// TODO: Implement regex matching for columns
			if t.caseSensitive {
				if !strings.Contains(colValue, colFilter) {
					match = false
					break
				}
			} else {
				if !strings.Contains(strings.ToLower(colValue), strings.ToLower(colFilter)) {
					match = false
					break
				}
			}
		} else {
			if t.caseSensitive {
				if !strings.Contains(colValue, colFilter) {
					match = false
					break
				}
			} else {
				if !strings.Contains(strings.ToLower(colValue), strings.ToLower(colFilter)) {
					match = false
					break
				}
			}
		}
	}

	if t.inverseFilter {
		match = !match
	}

	return match
}

// getResourceName returns the name of a resource
func (t *TUI) getResourceName(resource interface{}) string {
	switch r := resource.(type) {
	case v1.Pod:
		return r.Name
	case appsv1.Deployment:
		return r.Name
	case v1.Service:
		return r.Name
	case v1.ConfigMap:
		return r.Name
	case v1.Namespace:
		return r.Name
	default:
		return ""
	}
}

// getResourceColumnValue returns the value for a specific column of a resource
func (t *TUI) getResourceColumnValue(resource interface{}, colIndex int) string {
	switch r := resource.(type) {
	case v1.Pod:
		switch colIndex {
		case 0:
			return r.Name
		case 1:
			return string(r.Status.Phase)
		case 2:
			return t.getReadyCount(r)
		case 3:
			return t.formatDuration(time.Since(r.CreationTimestamp.Time))
		case 4:
			return r.Spec.NodeName
		}
	case appsv1.Deployment:
		switch colIndex {
		case 0:
			return r.Name
		case 1:
			return fmt.Sprintf("%d/%d", r.Status.ReadyReplicas, r.Status.Replicas)
		case 2:
			return fmt.Sprintf("%d", r.Status.UpdatedReplicas)
		case 3:
			return fmt.Sprintf("%d", r.Status.AvailableReplicas)
		case 4:
			return t.formatDuration(time.Since(r.CreationTimestamp.Time))
		}
	case v1.Service:
		switch colIndex {
		case 0:
			return r.Name
		case 1:
			return string(r.Spec.Type)
		case 2:
			return r.Spec.ClusterIP
		case 3:
			if len(r.Status.LoadBalancer.Ingress) > 0 {
				return r.Status.LoadBalancer.Ingress[0].IP
			}
			return "<none>"
		case 4:
			ports := ""
			for i, port := range r.Spec.Ports {
				if i > 0 {
					ports += ","
				}
				ports += fmt.Sprintf("%d", port.Port)
			}
			return ports
		}
	case v1.ConfigMap:
		switch colIndex {
		case 0:
			return r.Name
		case 1:
			return fmt.Sprintf("%d", len(r.Data)+len(r.BinaryData))
		case 2:
			return t.formatDuration(time.Since(r.CreationTimestamp.Time))
		}
	case v1.Namespace:
		switch colIndex {
		case 0:
			return r.Name
		case 1:
			return string(r.Status.Phase)
		case 2:
			return t.formatDuration(time.Since(r.CreationTimestamp.Time))
		}
	}
	return ""
}

// getTableHeaders returns table headers for the current resource type
func (t *TUI) getTableHeaders() []string {
	switch t.currentView {
	case ResourcePods:
		return []string{"Name", "Status", "Ready", "Age", "Node"}
	case ResourceDeployments:
		return []string{"Name", "Ready", "Up-to-date", "Available", "Age"}
	case ResourceServices:
		return []string{"Name", "Type", "Cluster-IP", "External-IP", "Ports"}
	case ResourceConfigMaps:
		return []string{"Name", "Data", "Age"}
	case ResourceNamespaces:
		return []string{"Name", "Status", "Age"}
	default:
		return []string{"Name", "Status", "Age"}
	}
}

// getColumnWidths calculates column widths based on available space
func (t *TUI) getColumnWidths(totalWidth, numColumns int) []int {
	if numColumns == 0 {
		return []int{}
	}

	// Account for borders and separators: 3 chars per column (│ space content space)
	availableWidth := totalWidth - (numColumns*3 + 1) // +1 for final │
	if availableWidth < numColumns*10 {               // Minimum 10 chars per column
		availableWidth = numColumns * 10
	}

	// Distribute width evenly
	widths := make([]int, numColumns)
	baseWidth := availableWidth / numColumns
	extra := availableWidth % numColumns

	for i := 0; i < numColumns; i++ {
		widths[i] = baseWidth
		if i < extra {
			widths[i]++
		}
	}

	return widths
}

// formatResourceLine formats a resource into a table line
func (t *TUI) formatResourceLine(resource interface{}, colWidths []int) string {
	headers := t.getTableHeaders()
	line := "│ "

	for i := range headers {
		value := t.getResourceColumnValue(resource, i)
		if len(value) > colWidths[i] {
			value = value[:colWidths[i]-3] + "..."
		}
		line += fmt.Sprintf("%-*s", colWidths[i], value)
		if i < len(headers)-1 {
			line += " │ "
		}
	}

	line += " │"
	return line
}

// drawStatusBar draws the status information bar
func (t *TUI) drawStatusBar(width, y int) {
	filtered := t.getFilteredResources()
	total := t.getCurrentViewCount()

	// Build status components
	namespaceInfo := fmt.Sprintf("📁 %s", t.namespace)
	resourceInfo := fmt.Sprintf("🎯 %s: %d/%d", t.currentView.DisplayName(), len(filtered), total)
	viewModeInfo := fmt.Sprintf("👁️ %s", t.getViewModeName())

	// Add filter info if active
	var filterInfo string
	if t.filter != "" || t.filterMode {
		filterInfo = fmt.Sprintf(" | 🔍 '%s'", t.filter)
	}

	// Combine status parts
	status := fmt.Sprintf("%s | %s | %s%s", namespaceInfo, resourceInfo, viewModeInfo, filterInfo)

	// Truncate if too long
	if len(status) > width-2 {
		status = status[:width-5] + "..."
	}

	// Pad to full width
	if len(status) < width {
		status += strings.Repeat(" ", width-len(status))
	}

	// Enhanced styling with gradient-like effect
	style := tcell.StyleDefault.Background(t.theme.accent).Foreground(tcell.ColorBlack).Bold(true)
	t.drawText(0, y, width, status, style)
}

// getViewModeName returns a display name for the current view mode
func (t *TUI) getViewModeName() string {
	switch t.viewMode {
	case ViewModeList:
		return "List"
	case ViewModeDetails:
		return "Details"
	case ViewModeYAML:
		return "YAML"
	case ViewModeLogs:
		return "Logs"
	case ViewModeRelationships:
		return "Relationships"
	default:
		return "Unknown"
	}
}

// getCurrentViewCount returns the total count for the current view
func (t *TUI) getCurrentViewCount() int {
	switch t.currentView {
	case ResourcePods:
		return len(t.pods)
	case ResourceDeployments:
		return len(t.deployments)
	case ResourceServices:
		return len(t.services)
	case ResourceConfigMaps:
		return len(t.configMaps)
	default:
		return 0
	}
}

// drawFooter draws the help/instruction footer
func (t *TUI) drawFooter(width, y int) {
	helpText := " ↑↓ Navigate │ Enter Details │ r Refresh │ d Delete │ c Create │ n Namespace │ / Search │ f Clear Filter │ h Help │ q Quit "
	if len(helpText) > width {
		helpText = helpText[:width-3] + "..."
	}

	style := tcell.StyleDefault.Background(tcell.ColorDarkGray).Foreground(tcell.ColorWhite)
	t.drawText(0, y, width, helpText, style)
}

// drawDetailsView draws the details view for selected resource
func (t *TUI) drawDetailsView(width, height int) {
	resource := t.getSelectedResource()
	if resource == nil {
		t.drawText(0, 0, width, "No resource selected", tcell.StyleDefault)
		return
	}

	// Header
	header := fmt.Sprintf(" 📋 %s Details ", t.currentView.DisplayName())
	t.drawText(0, 0, width, header, tcell.StyleDefault.Background(t.theme.header).Foreground(tcell.ColorWhite).Bold(true))

	// Details content
	details := t.getResourceDetails(resource)
	y := 2
	for _, line := range details {
		if y >= height-2 {
			break
		}
		t.drawText(0, y, width, line, tcell.StyleDefault)
		y++
	}

	// Footer
	footer := " ESC Back │ y YAML │ l Logs (pods only) "
	t.drawText(0, height-1, width, footer, tcell.StyleDefault.Background(t.theme.background).Foreground(t.theme.foreground))
}

// drawYAMLView draws the YAML view for selected resource
func (t *TUI) drawYAMLView(width, height int) {
	resource := t.getSelectedResource()
	if resource == nil {
		t.drawText(0, 0, width, "No resource selected", tcell.StyleDefault)
		return
	}

	// Header
	header := fmt.Sprintf(" 📄 %s YAML ", t.currentView.DisplayName())
	t.drawText(0, 0, width, header, tcell.StyleDefault.Background(t.theme.header).Foreground(tcell.ColorWhite).Bold(true))

	// YAML content
	yaml := t.getResourceYAML(resource)
	lines := strings.Split(yaml, "\n")

	y := 2
	for i := t.detailsScroll; i < len(lines) && y < height-2; i++ {
		line := lines[i]
		if len(line) > width {
			line = line[:width-3] + "..."
		}
		t.drawText(0, y, width, line, tcell.StyleDefault)
		y++
	}

	// Footer
	footer := " ESC Back │ ↑↓ Scroll "
	t.drawText(0, height-1, width, footer, tcell.StyleDefault.Background(t.theme.background).Foreground(t.theme.foreground))
}

// drawLogsView draws the logs view for selected pod
func (t *TUI) drawLogsView(width, height int) {
	if t.currentView != ResourcePods {
		t.drawText(0, 0, width, "Logs only available for pods", tcell.StyleDefault)
		return
	}

	resource := t.getSelectedResource()
	pod, ok := resource.(v1.Pod)
	if !ok {
		t.drawText(0, 0, width, "No pod selected", tcell.StyleDefault)
		return
	}

	// Header
	header := fmt.Sprintf(" 📋 Pod Logs: %s ", pod.Name)
	t.drawText(0, 0, width, header, tcell.StyleDefault.Background(t.theme.header).Foreground(tcell.ColorWhite).Bold(true))

	// Logs content (placeholder for now)
	logs := []string{
		"Log streaming not yet implemented...",
		"Use 'kubectl logs' command for now.",
		"",
		fmt.Sprintf("Pod: %s", pod.Name),
		fmt.Sprintf("Namespace: %s", pod.Namespace),
		fmt.Sprintf("Status: %s", pod.Status.Phase),
	}

	y := 2
	for i := t.logsScroll; i < len(logs) && y < height-2; i++ {
		line := logs[i]
		if len(line) > width {
			line = line[:width-3] + "..."
		}
		t.drawText(0, y, width, line, tcell.StyleDefault)
		y++
	}

	// Footer
	footer := " ESC Back │ ↑↓ Scroll "
	t.drawText(0, height-1, width, footer, tcell.StyleDefault.Background(t.theme.background).Foreground(t.theme.foreground))
}

// drawRelationshipsView draws the relationships view showing resource connections
func (t *TUI) drawRelationshipsView(width, height int) {
	// Header
	header := " 🔗 Resource Relationships "
	t.drawText(0, 0, width, header, tcell.StyleDefault.Background(t.theme.header).Foreground(tcell.ColorWhite).Bold(true))

	// Get all relationships
	relationships := t.getResourceRelationships()

	if len(relationships) == 0 {
		t.drawText(0, 2, width, "No relationships found", tcell.StyleDefault)
		return
	}

	// Display relationships
	y := 2
	for i := t.relationshipsScroll; i < len(relationships) && y < height-2; i++ {
		rel := relationships[i]
		line := fmt.Sprintf("%s → %s (%s)", rel.From, rel.To, rel.RelationType)
		if len(line) > width {
			line = line[:width-3] + "..."
		}
		t.drawText(0, y, width, line, tcell.StyleDefault)
		y++
	}

	// Footer
	footer := " ESC Back │ ↑↓ Scroll "
	t.drawText(0, height-1, width, footer, tcell.StyleDefault.Background(t.theme.background).Foreground(t.theme.foreground))
}

// getSelectedResource returns the currently selected resource
func (t *TUI) getSelectedResource() interface{} {
	filtered := t.getFilteredResources()
	if t.selected < 0 || t.selected >= len(filtered) {
		return nil
	}
	return filtered[t.selected]
}

// getResourceDetails returns formatted details for a resource
func (t *TUI) getResourceDetails(resource interface{}) []string {
	switch r := resource.(type) {
	case v1.Pod:
		return t.getPodDetails(r)
	case appsv1.Deployment:
		return t.getDeploymentDetails(r)
	case v1.Service:
		return t.getServiceDetails(r)
	case v1.ConfigMap:
		return t.getConfigMapDetails(r)
	}
	return []string{"Unknown resource type"}
}

// getResourceYAML returns YAML representation of a resource
func (t *TUI) getResourceYAML(resource interface{}) string {
	data, err := json.MarshalIndent(resource, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error marshaling YAML: %v", err)
	}
	return string(data)
}

// getResourceRelationships returns all resource relationships
func (t *TUI) getResourceRelationships() []Relationship {
	var relationships []Relationship

	// Pod relationships
	relationships = append(relationships, t.getPodRelationships()...)

	// Deployment relationships
	relationships = append(relationships, t.getDeploymentRelationships()...)

	// Service relationships
	relationships = append(relationships, t.getServiceRelationships()...)

	// ConfigMap relationships
	relationships = append(relationships, t.getConfigMapRelationships()...)

	return relationships
}

// getPodRelationships returns relationships for pods
func (t *TUI) getPodRelationships() []Relationship {
	var relationships []Relationship

	for _, pod := range t.pods {
		// Pod to Deployment relationship (via owner references)
		for _, owner := range pod.OwnerReferences {
			if owner.Kind == "Deployment" {
				relationships = append(relationships, Relationship{
					From:         pod.Name,
					To:           owner.Name,
					RelationType: "owned-by",
				})
			}
		}

		// Pod to Service relationship (via selectors)
		for _, svc := range t.services {
			if t.podMatchesService(pod, svc) {
				relationships = append(relationships, Relationship{
					From:         pod.Name,
					To:           svc.Name,
					RelationType: "exposed-by",
				})
			}
		}
	}

	return relationships
}

// getDeploymentRelationships returns relationships for deployments
func (t *TUI) getDeploymentRelationships() []Relationship {
	var relationships []Relationship

	for _, dep := range t.deployments {
		// Find pods owned by this deployment
		for _, pod := range t.pods {
			for _, owner := range pod.OwnerReferences {
				if owner.Kind == "Deployment" && owner.Name == dep.Name {
					relationships = append(relationships, Relationship{
						From:         dep.Name,
						To:           pod.Name,
						RelationType: "owns",
					})
				}
			}
		}
	}

	return relationships
}

// getServiceRelationships returns relationships for services
func (t *TUI) getServiceRelationships() []Relationship {
	var relationships []Relationship

	for _, svc := range t.services {
		// Find pods exposed by this service
		for _, pod := range t.pods {
			if t.podMatchesService(pod, svc) {
				relationships = append(relationships, Relationship{
					From:         svc.Name,
					To:           pod.Name,
					RelationType: "exposes",
				})
			}
		}
	}

	return relationships
}

// getConfigMapRelationships returns relationships for configmaps
func (t *TUI) getConfigMapRelationships() []Relationship {
	var relationships []Relationship

	// ConfigMaps are referenced by pods via volumes or env
	// This is a simplified implementation
	for _, cm := range t.configMaps {
		// Check if any pod references this configmap
		for _, pod := range t.pods {
			if t.podUsesConfigMap(pod, cm) {
				relationships = append(relationships, Relationship{
					From:         pod.Name,
					To:           cm.Name,
					RelationType: "uses",
				})
			}
		}
	}

	return relationships
}

// podMatchesService checks if a pod matches a service selector
func (t *TUI) podMatchesService(pod v1.Pod, svc v1.Service) bool {
	for key, value := range svc.Spec.Selector {
		if pod.Labels[key] != value {
			return false
		}
	}
	return true
}

// podUsesConfigMap checks if a pod uses a configmap
func (t *TUI) podUsesConfigMap(pod v1.Pod, cm v1.ConfigMap) bool {
	// Check volumes
	for _, volume := range pod.Spec.Volumes {
		if volume.ConfigMap != nil && volume.ConfigMap.Name == cm.Name {
			return true
		}
	}

	// Check environment variables
	for _, container := range pod.Spec.Containers {
		for _, env := range container.Env {
			if env.ValueFrom != nil && env.ValueFrom.ConfigMapKeyRef != nil && env.ValueFrom.ConfigMapKeyRef.Name == cm.Name {
				return true
			}
		}
	}

	return false
}

// getPodDetails returns formatted details for a pod
func (t *TUI) getPodDetails(pod v1.Pod) []string {
	return []string{
		fmt.Sprintf("Name: %s", pod.Name),
		fmt.Sprintf("Namespace: %s", pod.Namespace),
		fmt.Sprintf("Status: %s", pod.Status.Phase),
		fmt.Sprintf("Node: %s", pod.Spec.NodeName),
		fmt.Sprintf("Created: %s", pod.CreationTimestamp.Format("2006-01-02 15:04:05")),
		"",
		"Containers:",
	}
}

// getDeploymentDetails returns formatted details for a deployment
func (t *TUI) getDeploymentDetails(dep appsv1.Deployment) []string {
	return []string{
		fmt.Sprintf("Name: %s", dep.Name),
		fmt.Sprintf("Namespace: %s", dep.Namespace),
		fmt.Sprintf("Replicas: %d", dep.Status.Replicas),
		fmt.Sprintf("Ready: %d", dep.Status.ReadyReplicas),
		fmt.Sprintf("Available: %d", dep.Status.AvailableReplicas),
		fmt.Sprintf("Updated: %d", dep.Status.UpdatedReplicas),
		fmt.Sprintf("Created: %s", dep.CreationTimestamp.Format("2006-01-02 15:04:05")),
	}
}

// getServiceDetails returns formatted details for a service
func (t *TUI) getServiceDetails(svc v1.Service) []string {
	return []string{
		fmt.Sprintf("Name: %s", svc.Name),
		fmt.Sprintf("Namespace: %s", svc.Namespace),
		fmt.Sprintf("Type: %s", svc.Spec.Type),
		fmt.Sprintf("Cluster IP: %s", svc.Spec.ClusterIP),
		fmt.Sprintf("Created: %s", svc.CreationTimestamp.Format("2006-01-02 15:04:05")),
		"",
		"Ports:",
	}
}

// getConfigMapDetails returns formatted details for a configmap
func (t *TUI) getConfigMapDetails(cm v1.ConfigMap) []string {
	details := []string{
		fmt.Sprintf("Name: %s", cm.Name),
		fmt.Sprintf("Namespace: %s", cm.Namespace),
		fmt.Sprintf("Data items: %d", len(cm.Data)),
		fmt.Sprintf("Binary data items: %d", len(cm.BinaryData)),
		fmt.Sprintf("Created: %s", cm.CreationTimestamp.Format("2006-01-02 15:04:05")),
		"",
		"Data keys:",
	}

	for key := range cm.Data {
		details = append(details, fmt.Sprintf("  - %s", key))
	}

	return details
}

// drawHelpScreen shows the help screen
func (t *TUI) drawHelpScreen(width, height int) {
	t.screen.Clear()

	title := " 🚀 Kubernetes Dashboard - Help "
	padding := (width - len(title)) / 2
	titleBar := strings.Repeat("═", padding) + title + strings.Repeat("═", width-padding-len(title))
	t.drawText(0, 0, width, titleBar, tcell.StyleDefault.Background(tcell.ColorDarkBlue).Foreground(tcell.ColorWhite).Bold(true))

	helpLines := []string{
		"",
		" Navigation:",
		"   ↑↓, j/k     Navigate through resources",
		"   Tab         Switch between resource types",
		"   1-4         Jump to: Pods, Deployments, Services, ConfigMaps",
		"   Enter       Show resource details",
		"",
		" View Modes:",
		"   v           Cycle view modes (List → Details → YAML → Logs → Relationships)",
		"   y           YAML view",
		"   l           Logs view (pods only)",
		"   r           Relationships view",
		"",
		" Split Pane:",
		"   s           Toggle split-pane mode",
		"   S           Switch split layout (vertical/horizontal)",
		"",
		" Actions:",
		"   r, F5       Refresh all resources",
		"   d           Delete selected resource",
		"   c           Create new resource",
		"   n           Change namespace",
		"",
		" Search & Filter:",
		"   /           Search resources by name",
		"   f           Clear current filter",
		"",
		" General:",
		"   ?, h        Show this help",
		"   t, T        Cycle through color themes",
		"   q, Esc      Quit application",
		"",
		" Status Colors:",
		"   🟢 Green    Running/Ready",
		"   🟡 Yellow   Pending",
		"   🔴 Red      Failed/Error",
		"   🔵 Blue     Succeeded/Complete",
		"",
		" Press any key to return...",
	}

	y := 2
	for _, line := range helpLines {
		if y >= height-1 {
			break
		}
		t.drawText(0, y, width, line, tcell.StyleDefault)
		y++
	}
}

// drawLoadingScreen shows a loading screen
func (t *TUI) drawLoadingScreen(width, height int) {
	t.screen.Clear()

	loadingText := " 🔄 Loading Kubernetes resources..."
	y := height / 2
	x := (width - len(loadingText)) / 2

	t.drawText(x, y, width-x, loadingText, tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorYellow).Bold(true))
}

// formatPodLine formats a pod into a table line
func (t *TUI) formatPodLine(pod v1.Pod) string {
	name := pod.Name
	if len(name) > 24 {
		name = name[:21] + "..."
	}
	name = fmt.Sprintf("%-24s", name)

	status := string(pod.Status.Phase)
	status = fmt.Sprintf("%-11s", status)

	readyContainers := 0
	totalContainers := len(pod.Status.ContainerStatuses)
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Ready {
			readyContainers++
		}
	}
	ready := fmt.Sprintf("%d/%d", readyContainers, totalContainers)
	ready = fmt.Sprintf("%-7s", ready)

	age := time.Since(pod.CreationTimestamp.Time)
	ageStr := t.formatDuration(age)
	ageStr = fmt.Sprintf("%-11s", ageStr)

	node := pod.Spec.NodeName
	if node == "" {
		node = "<none>"
	}
	if len(node) > 15 {
		node = node[:12] + "..."
	}

	return name + status + ready + ageStr + node
}

// formatDuration formats a duration into a human-readable string
func (t *TUI) formatDuration(d time.Duration) string {
	if d.Hours() >= 24 {
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%dd", days)
	}
	if d.Hours() >= 1 {
		return fmt.Sprintf("%.0fh", d.Hours())
	}
	if d.Minutes() >= 1 {
		return fmt.Sprintf("%.0fm", d.Minutes())
	}
	return fmt.Sprintf("%.0fs", d.Seconds())
}

// drawText draws text at the specified position
func (t *TUI) drawText(x, y, maxWidth int, text string, style tcell.Style) {
	runes := []rune(text)
	for i, r := range runes {
		if i >= maxWidth {
			break
		}
		t.screen.SetContent(x+i, y, r, nil, style)
	}
}

// showPodDetails shows detailed information about the selected pod
func (t *TUI) showPodDetails() {
	if len(t.pods) == 0 || t.selected >= len(t.pods) {
		return
	}

	pod := t.pods[t.selected]

	// Create a simple details view
	details := []string{
		fmt.Sprintf("Name: %s", pod.Name),
		fmt.Sprintf("Namespace: %s", pod.Namespace),
		fmt.Sprintf("Status: %s", pod.Status.Phase),
		fmt.Sprintf("Node: %s", pod.Spec.NodeName),
		fmt.Sprintf("Created: %s", pod.CreationTimestamp.Format("2006-01-02 15:04:05")),
		"",
		"Containers:",
	}

	for _, container := range pod.Spec.Containers {
		details = append(details, fmt.Sprintf("  - %s: %s", container.Name, container.Image))
	}

	// Simple modal-like display (just overwrite the screen)
	t.screen.Clear()
	width, height := t.screen.Size()

	for i, line := range details {
		if i >= height-2 {
			break
		}
		t.drawText(0, i, width, line, tcell.StyleDefault)
	}

	t.drawText(0, height-1, width, "Press any key to return...", tcell.StyleDefault.Background(tcell.ColorRed).Foreground(tcell.ColorWhite))
	t.screen.Show()

	// Wait for any key press
	for {
		event := t.screen.PollEvent()
		if _, ok := event.(*tcell.EventKey); ok {
			break
		}
	}
}

// deleteSelectedPod deletes the currently selected pod
func (t *TUI) deleteSelectedPod() {
	if len(t.pods) == 0 || t.selected >= len(t.pods) {
		return
	}

	pod := t.pods[t.selected]

	// Show confirmation
	confirmMsg := fmt.Sprintf("Delete pod '%s'? (y/N)", pod.Name)
	t.drawText(0, 1, 50, confirmMsg, tcell.StyleDefault.Background(tcell.ColorYellow).Foreground(tcell.ColorBlack))
	t.screen.Show()

	// Wait for confirmation
	event := t.screen.PollEvent()
	if ev, ok := event.(*tcell.EventKey); ok && ev.Rune() == 'y' {
		err := k8s.DeletePod(t.clientset, pod.Namespace, pod.Name)
		if err != nil {
			klog.Errorf("Failed to delete pod: %v", err)
			errorMsg := fmt.Sprintf("Error deleting pod: %v", err)
			t.drawText(0, 3, 80, errorMsg, tcell.StyleDefault.Background(tcell.ColorRed).Foreground(tcell.ColorWhite))
			t.screen.Show()
			time.Sleep(2 * time.Second)
		} else {
			// Reload pods
			t.loadPods()
		}
	}
}

// changeNamespace allows changing the current namespace
func (t *TUI) changeNamespace() {
	// Fetch available namespaces
	namespaces, err := k8s.ListNamespaces(t.clientset)
	if err != nil {
		// Show error message
		t.screen.Clear()
		t.drawText(0, 0, 80, "Error: Failed to list namespaces: "+err.Error(), tcell.StyleDefault.Background(tcell.ColorRed).Foreground(tcell.ColorWhite))
		t.drawText(0, 2, 80, "Press any key to continue...", tcell.StyleDefault)
		t.screen.Show()
		t.screen.PollEvent()
		return
	}

	if len(namespaces) == 0 {
		// Show error message
		t.screen.Clear()
		t.drawText(0, 0, 80, "Error: No namespaces found", tcell.StyleDefault.Background(tcell.ColorRed).Foreground(tcell.ColorWhite))
		t.drawText(0, 2, 80, "Press any key to continue...", tcell.StyleDefault)
		t.screen.Show()
		t.screen.PollEvent()
		return
	}

	// Create list of namespace names
	var namespaceNames []string
	for _, ns := range namespaces {
		namespaceNames = append(namespaceNames, ns.Name)
	}

	// Simple selection dialog
	selectedIndex := 0
	for {
		t.screen.Clear()

		t.drawText(0, 0, 80, "Select Namespace (↑↓ to navigate, Enter to select, Esc to cancel):", tcell.StyleDefault.Bold(true))

		// Show namespaces
		for i, name := range namespaceNames {
			style := tcell.StyleDefault
			if i == selectedIndex {
				style = style.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite).Bold(true)
			}
			prefix := "  "
			if i == selectedIndex {
				prefix = "▶ "
			}
			t.drawText(0, i+2, 80, prefix+name, style)
		}

		t.screen.Show()

		event := t.screen.PollEvent()
		switch ev := event.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEnter:
				newNamespace := namespaceNames[selectedIndex]
				if newNamespace != t.namespace {
					t.namespace = newNamespace
					t.refreshData()
				}
				return
			case tcell.KeyEscape:
				return
			case tcell.KeyUp:
				if selectedIndex > 0 {
					selectedIndex--
				}
			case tcell.KeyDown:
				if selectedIndex < len(namespaceNames)-1 {
					selectedIndex++
				}
			}
		}
	}
}

// getReadyCount returns the ready container count as a string
func (t *TUI) getReadyCount(pod v1.Pod) string {
	readyContainers := 0
	totalContainers := len(pod.Status.ContainerStatuses)
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Ready {
			readyContainers++
		}
	}
	return fmt.Sprintf("%d/%d", readyContainers, totalContainers)
}

// searchDialog shows a search input dialog
func (t *TUI) searchDialog() {
	input := t.filter
	cursor := len(input)

	for {
		t.screen.Clear()

		prompt := "Search pods (current: " + t.filter + "): " + input
		t.drawText(0, 0, 80, prompt, tcell.StyleDefault)

		// Show cursor
		if cursor < len(input) {
			t.screen.SetContent(len(prompt)-len(input)+cursor, 0, '_', nil, tcell.StyleDefault)
		} else {
			t.drawText(len(prompt), 0, 1, "_", tcell.StyleDefault)
		}

		t.screen.Show()

		event := t.screen.PollEvent()
		switch ev := event.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEnter:
				t.filter = input
				t.selected = 0 // Reset selection
				return
			case tcell.KeyEscape:
				return
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				if len(input) > 0 {
					input = input[:len(input)-1]
					cursor--
					if cursor < 0 {
						cursor = 0
					}
				}
			case tcell.KeyRune:
				input += string(ev.Rune())
				cursor++
			}
		}
	}
}

// clearFilter clears the current search filter
func (t *TUI) clearFilter() {
	t.filter = ""
	t.selected = 0
}

// createPodDialog shows a simple pod creation dialog
func (t *TUI) createPodDialog() {
	name := ""
	image := "nginx:latest"
	cursor := 0
	field := 0 // 0 = name, 1 = image

	for {
		t.screen.Clear()

		lines := []string{
			"Create New Pod",
			"",
			fmt.Sprintf("Name: %s%s", name, t.getCursorText(field == 0, cursor, len(name))),
			fmt.Sprintf("Image: %s%s", image, t.getCursorText(field == 1, cursor, len(image))),
			"",
			"Tab: Switch field | Enter: Create | Esc: Cancel",
		}

		for i, line := range lines {
			t.drawText(0, i, 80, line, tcell.StyleDefault)
		}

		t.screen.Show()

		event := t.screen.PollEvent()
		switch ev := event.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEnter:
				if name != "" && image != "" {
					t.createPod(name, image)
				}
				return
			case tcell.KeyEscape:
				return
			case tcell.KeyTab:
				field = (field + 1) % 2
				cursor = 0
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				if field == 0 && len(name) > 0 {
					name = name[:len(name)-1]
					cursor--
				} else if field == 1 && len(image) > 0 {
					image = image[:len(image)-1]
					cursor--
				}
				if cursor < 0 {
					cursor = 0
				}
			case tcell.KeyRune:
				if field == 0 {
					name += string(ev.Rune())
					cursor++
				} else {
					image += string(ev.Rune())
					cursor++
				}
			}
		}
	}
}

// getCursorText returns cursor text for input fields
func (t *TUI) getCursorText(active bool, cursor, length int) string {
	if !active {
		return ""
	}
	if cursor < length {
		return "_" // Cursor in middle
	}
	return "_" // Cursor at end
}

// createPod creates a new pod with the given name and image
func (t *TUI) createPod(name, image string) {
	t.loading = true
	t.draw()
	t.screen.Show()

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: t.namespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  name,
					Image: image,
					Ports: []v1.ContainerPort{
						{
							ContainerPort: 80,
						},
					},
				},
			},
		},
	}

	_, err := k8s.CreatePod(t.clientset, t.namespace, pod)
	t.loading = false

	if err != nil {
		klog.Errorf("Failed to create pod: %v", err)
		errorMsg := fmt.Sprintf("Error creating pod: %v", err)
		t.drawText(0, 3, 80, errorMsg, tcell.StyleDefault.Background(tcell.ColorRed).Foreground(tcell.ColorWhite))
		t.screen.Show()
		time.Sleep(3 * time.Second)
	} else {
		// Reload pods
		t.loadPods()
	}
}
