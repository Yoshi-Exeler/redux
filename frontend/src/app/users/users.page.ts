import { Component, OnInit } from "@angular/core";
import { FormBuilder } from "@angular/forms";
import { AlertController, NavController, ModalController, IonRouterOutlet  } from "@ionic/angular";
import { API, User } from "../api/api";
import { UserAddModalComponent } from "../user-add-modal/user-add-modal.component";

@Component({
  selector: "app-users",
  templateUrl: "users.page.html",
  styleUrls: ["users.page.scss"],
})
export class UsersPage implements OnInit {
  users: User[];

  constructor(
    private navcontroller: NavController,
    private alertController: AlertController,
    private modalController: ModalController,
    private routerOutlet: IonRouterOutlet,
    private formBuilder: FormBuilder
    
  ) {
    this.getData();
  }

  ngOnInit(): void {
    this.getData();
  }

  async getData(ev?: any) {
    let ul = await API.GetUsers();
    this.users = ul.Users;
    if (ev != undefined) {
      ev.complete();
    }
  }

  async presentModal() {
    const modal = await this.modalController.create({
      component: UserAddModalComponent,
      componentProps: {reference: this},
      swipeToClose: true,
      cssClass: "auto-height",
      backdropDismiss: true,
      showBackdrop: true,
      breakpoints: [0],
      presentingElement: this.routerOutlet.nativeEl
    });
    return await modal.present();
  }

  async addUser(username: string, password: string, admin: boolean) {
    let ul = await API.AddUser(username, password, admin);
    this.users = ul.Users;
  }

  onLogout() {
    localStorage.setItem("authToken", "");
    this.navcontroller.navigateRoot("/redux/login");
  }

  async presentAlertConfirmDelete(user: User) {
    const alert = await this.alertController.create({
      header: "Confirm Deletion",
      message:
        "Are you sure you want to delete the user " + user.Username + " ?",
      buttons: [
        {
          text: "Cancel",
          role: "cancel",
          cssClass: "secondary",
          id: "cancel-button",
        },
        {
          text: "Confirm",
          id: "confirm-button",
          handler: () => {
            this.deleteUser(user);
          },
        },
      ],
    });

    await alert.present();
  }

  async deleteUser(user: User) {
    let ul = await API.DeleteUser(user.ID);
    this.users = ul.Users;
  }
}
