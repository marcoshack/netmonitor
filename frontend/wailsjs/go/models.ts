export namespace config {
	
	export class Settings {
	    test_interval_seconds: number;
	    data_retention_days: number;
	    notifications_enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.test_interval_seconds = source["test_interval_seconds"];
	        this.data_retention_days = source["data_retention_days"];
	        this.notifications_enabled = source["notifications_enabled"];
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
	    thresholds?: Thresholds;
	
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
	export class Config {
	    regions: Record<string, Region>;
	    settings?: Settings;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.regions = this.convertValues(source["regions"], Region, true);
	        this.settings = this.convertValues(source["settings"], Settings);
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
	
	
	

}

export namespace main {
	
	export class MonitoringStatus {
	    running: boolean;
	    lastTestTime: string;
	    nextTestTime: string;
	    totalEndpoints: number;
	
	    static createFrom(source: any = {}) {
	        return new MonitoringStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.running = source["running"];
	        this.lastTestTime = source["lastTestTime"];
	        this.nextTestTime = source["nextTestTime"];
	        this.totalEndpoints = source["totalEndpoints"];
	    }
	}
	export class SystemInfo {
	    applicationName: string;
	    version: string;
	    buildTime: string;
	    running: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SystemInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.applicationName = source["applicationName"];
	        this.version = source["version"];
	        this.buildTime = source["buildTime"];
	        this.running = source["running"];
	    }
	}

}

export namespace storage {
	
	export class TestResult {
	    // Go type: time
	    timestamp: any;
	    endpoint_id: string;
	    protocol: string;
	    status: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new TestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.endpoint_id = source["endpoint_id"];
	        this.protocol = source["protocol"];
	        this.status = source["status"];
	        this.error = source["error"];
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

}

