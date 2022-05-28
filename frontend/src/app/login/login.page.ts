import { Component, OnInit } from "@angular/core";
import { FormBuilder, FormGroup, Validators } from "@angular/forms";
import { NavController } from "@ionic/angular";
import { API } from "../api/api";

@Component({
  selector: "app-tab2",
  templateUrl: "login.page.html",
  styleUrls: ["login.page.scss"],
})
export class LoginPage implements OnInit {
  constructor(
    private formBuilder: FormBuilder,
    private navcontroller: NavController
  ) {}

  fgLogin: FormGroup;

  ngOnInit(): void {
    this.fgLogin = this.formBuilder.group({
      username: [null, [Validators.required]],
      password: [null, [Validators.required]],
    });
  }

  onLogin() {
    API.Authenticate(
      this.fgLogin.value.username,
      this.fgLogin.value.password
    ).then((resp) => {
      if (resp.status === 200) {
        this.navcontroller.navigateRoot("/redux/files");
      }
    });
  }
}
