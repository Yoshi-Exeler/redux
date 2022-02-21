import { Component } from '@angular/core';
import { NgxFileDropEntry } from 'ngx-file-drop';
import { API, File, Folder, FolderContent } from '../api/api';

@Component({
  selector: 'app-tab1',
  templateUrl: 'tab1.page.html',
  styleUrls: ['tab1.page.scss']
})
export class Tab1Page {

  files: File[] = [];
  folders: Folder[] = [];
  path: string = "";
  constructor() {
    this.getData();
  }

  openFileContext(file: File) {
    console.log("Placeholder for file action", file);
  }

  async onDownloadFile(file: File) {
    // make an api call to fetch the file content
    let blob = await API.GetFileContent(file.Path);
    this.downloadBase64File("application/octet-stream", blob.Blob, file.Name);
  }

  downloadBase64File(contentType: string, base64Data: string, fileName: string) {
    const linkSource = `data:${contentType};base64,${base64Data}`;
    const downloadLink = document.createElement("a");
    downloadLink.href = linkSource;
    downloadLink.download = fileName;
    downloadLink.click();
  }

  // Function to download data to a file
  download(data, filename, type) {
    var file = new Blob([data], { type: type });
    var a = document.createElement("a"),
      url = URL.createObjectURL(file);
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    setTimeout(function () {
      document.body.removeChild(a);
      window.URL.revokeObjectURL(url);
    }, 0);
  }

  canNavigateBack(): boolean {
    return this.path.length > 0
  }

  navigateIntoFolder(folder: Folder) {
    this.path += folder.Name + "/"
    console.log("Navigate to ", this.path);
    this.getData();
  }

  async onFileUpload(f: FileList) {
    // iterate over the selected files
    for (let i = 0; i < f.length; i++) {
      // read the content of the current file
      const reader = new FileReader();
      let value = null;
      reader.addEventListener('load', (event) => {
        value = event.target.result;
        // upload the current file
        API.UploadFile(this.path + f[i].name, window.btoa(value), this.path).then((resp: FolderContent) => {
          this.files = resp.Files;
          this.folders = resp.Folders;
        })
        console.log("RAW:", value)
      });

      reader.readAsBinaryString(f[i]);

    }
  }

  navigateBack() {
    let pathSegments = this.path.split("/");
    if (pathSegments.length === 2) {
      this.path = "";
      console.log("Navigate back to origin ", this.path)
      this.getData();
      return;
    }
    // navigate back to the last Segment
    let trunc = this.path.substring(0, this.path.lastIndexOf("/"));
    let withoutSegment = trunc.substring(0, trunc.lastIndexOf("/")) + "/";
    this.path = withoutSegment;
    console.log("Navigate back to path ", this.path)
    this.getData();
  }

  getData() {
    API.GetFolderContent(this.path).then((resp) => {
      this.files = resp.Files;
      this.folders = resp.Folders;
    })
  }

  getIconForType(type: string): string {
    // video types
    if (
      type === "mp4" ||
      type === "mov" ||
      type === "webm") {
      return "videocam-outline";
    }
    // text types
    if (
      type === "txt" ||
      type === "odt" ||
      type === "pdf") {
      return "document-text-outline";
    }
    return "document-outline";
  }

}

