export namespace logic {
	
	export class Account {
	    id: string;
	    alias: string;
	    username: string;
	    password: string;
	    game_id: string;
	    token: string;
	    device_fingerprint: string;
	    is_first_login: boolean;
	    create_time: number;
	
	    static createFrom(source: any = {}) {
	        return new Account(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.alias = source["alias"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.game_id = source["game_id"];
	        this.token = source["token"];
	        this.device_fingerprint = source["device_fingerprint"];
	        this.is_first_login = source["is_first_login"];
	        this.create_time = source["create_time"];
	    }
	}
	export class ConfigData {
	    theme: string;
	    enabled_tags: string[];
	    accounts: Account[];
	    window_width: number;
	    window_height: number;
	    window_x: number;
	    window_y: number;
	    game_paths: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new ConfigData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.theme = source["theme"];
	        this.enabled_tags = source["enabled_tags"];
	        this.accounts = this.convertValues(source["accounts"], Account);
	        this.window_width = source["window_width"];
	        this.window_height = source["window_height"];
	        this.window_x = source["window_x"];
	        this.window_y = source["window_y"];
	        this.game_paths = source["game_paths"];
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

