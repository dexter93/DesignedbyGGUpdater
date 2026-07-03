export namespace main {
	
	export class Device {
	    vid: string;
	    pid: string;
	    name: string;
	    manufacturer: string;
	    product: string;
	    serialNumber: string;
	    path: string;
	    isBootloader: boolean;
	    firmwarePath: string;
	
	    static createFrom(source: any = {}) {
	        return new Device(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.vid = source["vid"];
	        this.pid = source["pid"];
	        this.name = source["name"];
	        this.manufacturer = source["manufacturer"];
	        this.product = source["product"];
	        this.serialNumber = source["serialNumber"];
	        this.path = source["path"];
	        this.isBootloader = source["isBootloader"];
	        this.firmwarePath = source["firmwarePath"];
	    }
	}
	export class FlashResult {
	    success: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new FlashResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	    }
	}

}

