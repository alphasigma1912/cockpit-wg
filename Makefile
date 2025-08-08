UI_DIR := ui
BRIDGE_DIR := bridge
DIST_DIR := dist/cockpit-wg

.PHONY: ui bridge dist clean

ui:
	npm --prefix $(UI_DIR) install
	npm --prefix $(UI_DIR) run build

bridge:
	mkdir -p $(DIST_DIR)
	cd $(BRIDGE_DIR) && CGO_ENABLED=0 go build -ldflags "-s -w" -o ../$(DIST_DIR)/wg-bridge

dist: ui bridge
	mkdir -p $(DIST_DIR)
	cp $(UI_DIR)/manifest.json $(DIST_DIR)/manifest.json
	cp $(UI_DIR)/dist/index.html $(DIST_DIR)/index.html
	cp -r $(UI_DIR)/dist/assets $(DIST_DIR)/assets

clean:
	rm -rf dist $(UI_DIR)/node_modules $(UI_DIR)/dist
