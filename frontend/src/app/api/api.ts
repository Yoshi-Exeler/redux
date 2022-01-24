import { JsonPipe } from "@angular/common";
import axios from "axios";

export class API {

    public static apiurl: string = "http://localhost:8080"

    static async GetFolderContent(path: string): Promise<FolderContent> {
        let promise = axios.post(this.apiurl + "/getfoldercontent", JSON.stringify({ path: path }));
        promise.catch((err) => { console.log("HTTP ERROR" + err); })
        promise.then((resp) => { console.log("HTTP-POST: OK >>" + JSON.stringify(resp.data)); })
        let result = await promise
        return result.data;
    }
}

export class FolderContent {
    Files: File[];
    Folders: Folder[]
}

export class Folder {
    Name: string;
    Path: string;
}

export class File {
    Name: string;
    Extension: string;
    Path: string;
}
