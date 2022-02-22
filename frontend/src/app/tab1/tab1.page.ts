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
    this.getData(null);
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
    this.getData(null);
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
      this.getData(null);
      return;
    }
    // navigate back to the last Segment
    let trunc = this.path.substring(0, this.path.lastIndexOf("/"));
    let withoutSegment = trunc.substring(0, trunc.lastIndexOf("/")) + "/";
    this.path = withoutSegment;
    console.log("Navigate back to path ", this.path)
    this.getData(null);
  }

  getData(ev: any) {
    API.GetFolderContent(this.path).then((resp) => {
      this.files = resp.Files;
      this.folders = resp.Folders;
      if (ev != null) {
        ev.complete();
      }
    })
  }

  getIconForType(type: string): string {
    // videos
    if (this.in(type, ["mp4", "mov", "avi", "webm"])) {
      return "videocam-outline";
    }
    // pictures
    if (this.in(type, ["png", "jpg", "jpeg", "ico", "bmp"])) {
      return "image-outline"
    }
    // audio
    if (this.in(type, ["mp3", "wav", "ogg"])) {
      return "musical-notes-outline"
    }
    // config files
    if (this.in(type, ["json", "xml", "yaml", "yml", "toml", "tml", "ini"])) {
      return "cog-outline"
    }
    // executables / libraries
    if (this.in(type, ["exe", "dll", "so", "jar"])) {
      return "hardware-chip-outline"
    }
    // source code
    if (this.in(type, ["go", "js", "ts", "html", "css", "c", "cpp", "h", "rs", "py", "bat", "sh", "asm", "nasm", "java", "cs", "vbs"])) {
      return "code-slash-outline"
    }
    // text types
    if (this.in(type, ["txt", "odt", "doc", "docx"])) {
      return "document-text-outline"
    }
    // table files
    if (this.in(type, ["csv", "ods", "xlsx"])) {
      return "bar-chart-outline"
    }
    return "document-outline";
  }

  in(s: string, range: string[]): boolean {
    for (let i = 0; i < range.length; i++) {
      if (s === range[i]) {
        return true;
      }
    }
    return false;
  }

}

