export namespace domain {
	
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
	    releaseNotes: string;
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
	        this.releaseNotes = source["releaseNotes"];
	        this.downloadURL = source["downloadURL"];
	        this.installerName = source["installerName"];
	        this.assetSizeBytes = source["assetSizeBytes"];
	    }
	}

}

