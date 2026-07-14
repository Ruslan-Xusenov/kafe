










func printBulkEditItems(orderID int, addedItems, cancelledItems []interface{}, waiterName, tableNumber string) {
	// Group items by printer target
	targetsMap := make(map[string]struct{})
	
	// Default target if none specified
	for _, it := range addedItems {
		item := it.(map[string]interface{})
		t, _ := item["printer_target"].(string)
		if t == "" { t = "ALL" }
		targetsMap[t] = struct{}{}
	}
	for _, it := range cancelledItems {
		item := it.(map[string]interface{})
		t, _ := item["printer_target"].(string)
		if t == "" { t = "ALL" }
		targetsMap[t] = struct{}{}
	}

	targets := []string{"USB"}
	targets = append(targets, networkPrinters...)

	for _, target := range targets {
		var addedForTarget []interface{}
		var cancelledForTarget []interface{}

		for _, it := range addedItems {
			item := it.(map[string]interface{})
			t, _ := item["printer_target"].(string)
			if t == "" { t = "ALL" }
			if t == "ALL" || t == target {
				addedForTarget = append(addedForTarget, item)
			}
		}

		for _, it := range cancelledItems {
			item := it.(map[string]interface{})
			t, _ := item["printer_target"].(string)
			if t == "" { t = "ALL" }
			if t == "ALL" || t == target {
				cancelledForTarget = append(cancelledForTarget, item)
			}
		}

		if len(addedForTarget) == 0 && len(cancelledForTarget) == 0 {
			continue
		}

		generateAndPrintBulkEditReceipt(target, orderID, addedForTarget, cancelledForTarget, waiterName, tableNumber)
	}
}

func generateAndPrintBulkEditReceipt(target string, orderID int, addedItems, cancelledItems []interface{}, waiterName, tableNumber string) {
	f, err := os.CreateTemp("", fmt.Sprintf("bulk_%d_*.bin", orderID))
	if err != nil {
		return
	}
	defer os.Remove(f.Name())

	// Build Binary Sequence
	f.Write(ESC_INIT)
	f.Write(DISABLE_CHINESE)
	f.Write(CODE_PAGE)
	f.Write(BEEP)
	f.Write(BEEP)

	f.Write(ALIGN_CENTER)
	f.Write(FONT_BIG)
	f.Write(toCP866("ИЗМЕНЕНИЯ / O'ZGARISHLAR\n"))
	f.Write(FONT_NORMAL)
	f.Write(toCP866("------------------------------------------------\n"))
	
	f.Write(ALIGN_LEFT)
	f.Write(toCP866(fmt.Sprintf("Чек №: %d\n", orderID)))
	f.Write(toCP866(fmt.Sprintf("Стол: %s\n", tableNumber)))
	f.Write(toCP866(fmt.Sprintf("Официант: %s\n", waiterName)))
	f.Write(toCP866(fmt.Sprintf("Время: %s\n", time.Now().Format("02.01.2006 15:04:05"))))
	f.Write(toCP866("------------------------------------------------\n\n"))

	f.Write(FONT_DOUBLE_H)
	if len(addedItems) > 0 {
		f.Write(toCP866("[+] ДОБАВЛЕНО:\n"))
		for _, it := range addedItems {
			item := it.(map[string]interface{})
			name := item["product_name"].(string)
			qty := item["quantity"].(float64)
			f.Write(toCP866(fmt.Sprintf("  + %s x %.1f\n", name, qty)))
		}
		f.Write([]byte("\n"))
	}

	if len(cancelledItems) > 0 {
		f.Write(toCP866("[-] ОТМЕНЕНО:\n"))
		for _, it := range cancelledItems {
			item := it.(map[string]interface{})
			name := item["product_name"].(string)
			qty := item["quantity"].(float64)
			f.Write(toCP866(fmt.Sprintf("  - %s x %.1f\n", name, qty)))
		}
		f.Write([]byte("\n"))
	}
	
	f.Write(FONT_NORMAL)
	f.Write(toCP866("------------------------------------------------\n\n\n\n"))
	f.Write(PAPER_CUT)
	f.Close()

	if target == "USB" {
		exec.Command("cmd", "/c", "copy", "/b", f.Name(), printerDevice).Run()
	} else {
		data, err := os.ReadFile(f.Name())
		if err == nil {
			conn, err := net.DialTimeout("tcp", target, 5*time.Second)
			if err == nil {
				defer conn.Close()
				conn.Write(data)
			}
		}
	}
}
