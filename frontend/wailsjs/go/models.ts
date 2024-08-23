export namespace main {
	
	export class CustomProfile {
	    name: string;
	    username: string;
	    host: string;
	    port: string;
	    password: string;
	
	    static createFrom(source: any = {}) {
	        return new CustomProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.username = source["username"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.password = source["password"];
	    }
	}
	export class YAMLConfig {
	
	
	    static createFrom(source: any = {}) {
	        return new YAMLConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

