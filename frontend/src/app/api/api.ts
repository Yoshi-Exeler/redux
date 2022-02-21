import { JsonPipe } from "@angular/common";
import axios from "axios";

export class API {

    public static apiurl: string = "http://192.168.0.20:8080"

    static async GetFolderContent(path: string): Promise<FolderContent> {
        let promise = axios.post(this.apiurl + "/getfoldercontent", JSON.stringify({ path: path }));
        promise.catch((err) => { console.log("HTTP ERROR" + err); })
        promise.then((resp) => { console.log("HTTP-POST: OK >>" + JSON.stringify(resp.data)); })
        let result = await promise
        return result.data;
    }

    static async GetFileContent(path: string): Promise<FileContent> {
        let promise = axios.post(this.apiurl + "/getfilecontent", JSON.stringify({ path: path }));
        promise.catch((err) => { console.log("HTTP ERROR" + err); })
        promise.then((resp) => { console.log("HTTP-POST: OK >>" + JSON.stringify(resp.data)); })
        let result = await promise
        return result.data;
    }

    static async UploadFile(path: string, blob: string, dir: string): Promise<FolderContent> {
        let promise = axios.post(this.apiurl + "/fileupload", JSON.stringify({ path: path, blob: blob, currentDir: dir }));
        promise.catch((err) => { console.log("HTTP ERROR" + err); })
        promise.then((resp) => { console.log("HTTP-POST: OK >>" + JSON.stringify(resp.data)); })
        let result = await promise
        return result.data;
    }

}

export class FileContent {
    Blob: string;
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
