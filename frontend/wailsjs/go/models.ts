export namespace models {
	
	export class AppSettings {
	    test_interval_seconds: number;
	    data_retention_days: number;
	    notifications_enabled: boolean;
	    window_width?: number;
	    window_height?: number;
	    window_x?: number;
	    window_y?: number;
	
	    static createFrom(source: any = {}) {
	        return new AppSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.test_interval_seconds = source["test_interval_seconds"];
	        this.data_retention_days = source["data_retention_days"];
	        this.notifications_enabled = source["notifications_enabled"];
	        this.window_width = source["window_width"];
	        this.window_height = source["window_height"];
	        this.window_x = source["window_x"];
	        this.window_y = source["window_y"];
	    }
	}
	export class Thresholds {
	    latency_ms: number;
	    availability_percent: number;
	
	    static createFrom(source: any = {}) {
	        return new Thresholds(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.latency_ms = source["latency_ms"];
	        this.availability_percent = source["availability_percent"];
	    }
	}
	export class Endpoint {
	    name: string;
	    type: string;
	    address: string;
	    timeout: number;
	
	    static createFrom(source: any = {}) {
	        return new Endpoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.address = source["address"];
	        this.timeout = source["timeout"];
	    }
	}
	export class Region {
	    endpoints: Endpoint[];
	    thresholds: Thresholds;
	
	    static createFrom(source: any = {}) {
	        return new Region(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.endpoints = this.convertValues(source["endpoints"], Endpoint);
	        this.thresholds = this.convertValues(source["thresholds"], Thresholds);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Configuration {
	    regions: Record<string, Region>;
	    settings: AppSettings;
	
	    static createFrom(source: any = {}) {
	        return new Configuration(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.regions = this.convertValues(source["regions"], Region, true);
	        this.settings = this.convertValues(source["settings"], AppSettings);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class TestResult {
	    ts: number;
	    id: string;
	    ms: number;
	    st: number;
	    err: any;
	
	    static createFrom(source: any = {}) {
	        return new TestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ts = source["ts"];
	        this.id = source["id"];
	        this.ms = source["ms"];
	        this.st = source["st"];
	        this.err = source["err"];
	    }
	}

}

