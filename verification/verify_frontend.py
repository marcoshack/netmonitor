from playwright.sync_api import sync_playwright

def run(playwright):
    browser = playwright.chromium.launch(headless=True)
    page = browser.new_page()
    page.add_init_script("""
        window.go = {
            main: {
                App: {
                    GetConfig: async () => ({
                        regions: {
                            "Default": {
                                endpoints: [
                                    { name: "EP1", type: "HTTP", address: "http://test.com" }
                                ]
                            }
                        },
                        settings: {
                            test_interval_seconds: 300,
                            data_retention_days: 90,
                            notifications_enabled: true
                        }
                    }),
                    GetConfigWarnings: async () => ([
                        "Duplicate endpoint ignored: EP2 (HTTP:http://test.com) in region Default"
                    ]),
                    GetHistoryRange: async () => ([]),
                    RemoveDuplicateEndpoints: async () => ""
                }
            }
        };
        window.runtime = {
            EventsOn: (evt, cb) => {}
        };
    """)

    page.on("console", lambda msg: print(f"Console: {msg.text}"))
    page.goto("http://localhost:8000/frontend/index.html")

    # Check if module failed to load due to Chart.js not being found?
    # import Chart from 'chart.js/auto'; -> This expects node_modules resolution.
    # In browser serving static files, bare imports are not supported without importmap or bundler.
    # Since we are just serving files raw, the import will fail.

    # We need to simulate a bundled environment or use import maps, or just accept we can't fully run it without build.
    # But wait, the repo has ?
    # If  folder has node_modules, we might be able to resolve it if we serve it right?
    # But standard browser doesn't resolve 'chart.js/auto' to 'node_modules/chart.js/auto/auto.js'.

    # So the JS likely halts at the import.

    page.screenshot(path="verification/frontend_mock.png")
    browser.close()

with sync_playwright() as playwright:
    run(playwright)
