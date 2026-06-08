export namespace domain {
	
	export class ThemeColors {
	    background: string;
	    surface: string;
	    text: string;
	    accent: string;
	    accentLight: string;
	
	    static createFrom(source: any = {}) {
	        return new ThemeColors(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.background = source["background"];
	        this.surface = source["surface"];
	        this.text = source["text"];
	        this.accent = source["accent"];
	        this.accentLight = source["accentLight"];
	    }
	}
	export class Theme {
	    id: string;
	    name: string;
	    colors: ThemeColors;
	
	    static createFrom(source: any = {}) {
	        return new Theme(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.colors = this.convertValues(source["colors"], ThemeColors);
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
	
	export class Track {
	    id: number;
	    trackName: string;
	    artistName: string;
	    albumName: string;
	    duration: number;
	    instrumental: boolean;
	    plainLyrics: string;
	    syncedLyrics: string;
	
	    static createFrom(source: any = {}) {
	        return new Track(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.trackName = source["trackName"];
	        this.artistName = source["artistName"];
	        this.albumName = source["albumName"];
	        this.duration = source["duration"];
	        this.instrumental = source["instrumental"];
	        this.plainLyrics = source["plainLyrics"];
	        this.syncedLyrics = source["syncedLyrics"];
	    }
	}

}

export namespace main {
	
	export class UpdateResult {
	    available: boolean;
	    latestVersion: string;
	    downloadURL: string;
	    installerName: string;
	    assetSizeBytes: number;
	
	    static createFrom(source: any = {}) {
	        return new UpdateResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.available = source["available"];
	        this.latestVersion = source["latestVersion"];
	        this.downloadURL = source["downloadURL"];
	        this.installerName = source["installerName"];
	        this.assetSizeBytes = source["assetSizeBytes"];
	    }
	}

}

