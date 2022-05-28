import { JsonPipe } from "@angular/common";
import axios, { AxiosResponse } from "axios";

export class API {
  public static apiurl: string = "http://192.168.0.15:8050";
  public static token: string = "";

  static async GetFolderContent(path: string): Promise<FolderContent> {
    let promise = axios.post(
      this.apiurl + "/getfoldercontent",
      JSON.stringify({ path: path, token: localStorage.getItem('authToken') })
    );
    promise.catch((err) => {
      console.error("[API][ERROR]" + err);
    });
    promise.then((resp) => {
      console.log("[API][OKAY]" + JSON.stringify(resp.data));
    });
    let result = await promise;
    return result.data;
  }

  static async GetFileContent(path: string): Promise<FileContent> {
    let promise = axios.post(
      this.apiurl + "/getfilecontent",
      JSON.stringify({ path: path, token: localStorage.getItem('authToken') })
    );
    promise.catch((err) => {
      console.error("[API][ERROR]" + err);
    });
    promise.then((resp) => {
      console.log("[API][OKAY]" + JSON.stringify(resp.data));
    });
    let result = await promise;
    return result.data;
  }

  static async UploadFile(
    path: string,
    blob: string,
    dir: string
  ): Promise<FolderContent> {
    let promise = axios.post(
      this.apiurl + "/fileupload",
      JSON.stringify({
        path: path,
        blob: blob,
        currentDir: dir,
        token: localStorage.getItem('authToken'),
      })
    );
    promise.catch((err) => {
      console.error("[API][ERROR]" + err);
    });
    promise.then((resp) => {
      console.log("[API][OKAY]" + JSON.stringify(resp.data));
    });
    let result = await promise;
    return result.data;
  }

  static async Authenticate(
    username: string,
    password: string
  ): Promise<AxiosResponse<any, any>> {
    let promise = axios.post(
      this.apiurl + "/authenticate",
      JSON.stringify({
        username: username,
        password: password
      })
    );
    promise.catch((err) => {
      console.error("[API][ERROR]" + err);
    });
    promise.then((resp) => {
      console.log("[API][OKAY]" + JSON.stringify(resp.data));
      console.log(resp.data)
      localStorage.setItem('authToken', resp.data.Token);
    });
    return promise;
  }
}

export class FileContent {
  Blob: string;
}

export class FolderContent {
  Files: File[];
  Folders: Folder[];
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
