export namespace main {
	
	export class ConnectionResult {
	    name: string;
	    host: string;
	    port: string;
	    username: string;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.username = source["username"];
	    }
	}
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
	export class SSHConfig {
	    name: string;
	    host: string;
	    port: number;
	    username: string;
	    password: string;
	    ssh_key_path: string;
	
	    static createFrom(source: any = {}) {
	        return new SSHConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.ssh_key_path = source["ssh_key_path"];
	    }
	}

}

