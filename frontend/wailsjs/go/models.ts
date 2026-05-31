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

