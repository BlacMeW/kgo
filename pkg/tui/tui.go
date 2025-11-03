package tui

import (
	"fmt"
	"strings"
	"time"

	"k8s-dashboard/pkg/k8s"

	"github.com/gdamore/tcell/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

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
		namespace: "default",
		filter:    "",
		showHelp:  false,
		loading:   false,
	}, nil
}

// Run starts the TUI main loop
func (t *TUI) Run() error {
	defer t.screen.Fini()

	// Initial pod load
	if err := t.loadPods(); err != nil {
		return fmt.Errorf("failed to load pods: %v", err)
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

			switch ev.Key() {
			case tcell.KeyEscape, tcell.KeyCtrlC:
				return nil
			case tcell.KeyDown:
				t.moveSelection(1)
			case tcell.KeyUp:
				t.moveSelection(-1)
			case tcell.KeyEnter:
				t.showPodDetails()
			case tcell.KeyF5:
				t.loadPods()
			case tcell.KeyRune:
				switch ev.Rune() {
				case 'q':
					return nil
				case 'r':
					t.loadPods()
				case 'd':
					t.deleteSelectedPod()
				case 'n':
					t.changeNamespace()
				case 'c':
					t.createPodDialog()
				case 'h':
					t.showHelp = true
				case '/':
					t.searchDialog()
				case 'f':
					t.clearFilter()
				}
			}
		case *tcell.EventResize:
			t.screen.Sync()
		}
	}
}

// loadPods fetches pods from the current namespace
func (t *TUI) loadPods() error {
	t.loading = true
	t.draw()
	t.screen.Show()

	pods, err := k8s.ListPods(t.clientset, t.namespace)
	t.loading = false

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

// moveSelection moves the selection up or down
func (t *TUI) moveSelection(delta int) {
	if len(t.pods) == 0 {
		return
	}

	t.selected += delta
	if t.selected < 0 {
		t.selected = 0
	}
	if t.selected >= len(t.pods) {
		t.selected = len(t.pods) - 1
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

	// Draw title bar
	t.drawTitleBar(width)

	// Draw search bar if filter is active
	if t.filter != "" {
		t.drawSearchBar(width, 2)
	}

	// Draw main content area
	contentStartY := 2
	if t.filter != "" {
		contentStartY = 4
	}
	contentHeight := height - 6 // Reserve space for title, status, and footer
	t.drawPodTable(width, contentHeight, contentStartY)

	// Draw status bar
	t.drawStatusBar(width, height-3)

	// Draw footer with instructions
	t.drawFooter(width, height-1)
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

// drawStatusBar draws the status information bar
func (t *TUI) drawStatusBar(width, y int) {
	filteredPods := t.getFilteredPods()
	status := fmt.Sprintf(" 📁 %s | 🎯 %d/%d pods", t.namespace, len(filteredPods), len(t.pods))

	if t.filter != "" {
		status += fmt.Sprintf(" | 🔍 '%s'", t.filter)
	}

	// Pad to full width
	if len(status) < width {
		status += strings.Repeat(" ", width-len(status))
	}

	style := tcell.StyleDefault.Background(tcell.ColorDarkGreen).Foreground(tcell.ColorWhite)
	t.drawText(0, y, width, status, style)
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
		"   ↑↓          Navigate through pods",
		"   Enter       Show pod details",
		"",
		" Actions:",
		"   r, F5       Refresh pod list",
		"   d           Delete selected pod",
		"   c           Create new pod",
		"   n           Change namespace",
		"",
		" Search & Filter:",
		"   /           Search pods by name",
		"   f           Clear current filter",
		"",
		" General:",
		"   h           Show this help",
		"   q, Esc      Quit application",
		"",
		" Status Colors:",
		"   🟢 Green    Running",
		"   🟡 Yellow   Pending",
		"   🔴 Red      Failed",
		"   🔵 Blue     Succeeded",
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

	loadingText := " 🔄 Loading pods..."
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
	input := t.namespace
	cursor := len(input)

	for {
		t.screen.Clear()

		prompt := "Enter namespace (current: " + t.namespace + "): " + input
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
				if input != "" {
					t.namespace = input
					t.loadPods()
				}
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
