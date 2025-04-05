package pkg

// create_window_map creates a map of windows by ID for quick lookup
func create_window_map(windows []Window) map[string]Window {
	window_map := make(map[string]Window)
	for _, w := range windows {
		window_map[w.ID] = w
	}
	return window_map
}

// create_id_set creates a set of window IDs for quick existence checks
func create_id_set(windows []Window) map[string]bool {
	id_set := make(map[string]bool)
	for _, w := range windows {
		id_set[w.ID] = true
	}
	return id_set
}

// find_window_by_id finds a window with the given ID in the list
func find_window_by_id(windows []Window, id string) (Window, bool) {
	for _, w := range windows {
		if w.ID == id {
			return w, true
		}
	}
	return Window{}, false
}

// add_active_window adds the active window to the result list if it exists
func add_active_window(result *[]Window, windows []Window, active_id string) {
	if active_window, found := find_window_by_id(windows, active_id); found {
		*result = append(*result, active_window)
	}
}

// add_existing_windows adds windows from current_list that still exist in new_list
func add_existing_windows(result *[]Window, current_list []Window, new_map map[string]Window, active_id string) {
	for _, w := range current_list {
		if w.ID == active_id {
			continue // Skip active window (already added)
		}
		if new_window, exists := new_map[w.ID]; exists {
			*result = append(*result, new_window)
		}
	}
}

// add_new_windows adds windows that weren't in the current list
func add_new_windows(result *[]Window, new_list []Window, existing_ids map[string]bool, active_id string) {
	for _, w := range new_list {
		if !existing_ids[w.ID] && w.ID != active_id {
			*result = append(*result, w)
		}
	}
}

// merge_window_lists merges two window lists while preserving order
func merge_window_lists(current_list []Window, new_list []Window, active_id string) []Window {
	// Create lookup structures
	new_window_map := create_window_map(new_list)
	existing_ids := create_id_set(current_list)

	// Start with an empty result
	result := []Window{}

	// Build the merged list in three steps
	add_active_window(&result, new_list, active_id)
	add_existing_windows(&result, current_list, new_window_map, active_id)
	add_new_windows(&result, new_list, existing_ids, active_id)

	return result
}
