import { Component, OnInit } from "@angular/core";
import { AlertController, NavController } from "@ionic/angular";
import { API, User } from "../api/api";

@Component({
  selector: "app-users",
  templateUrl: "users.page.html",
  styleUrls: ["users.page.scss"],
})
export class UsersPage implements OnInit {
  users: User[];

  constructor(private navcontroller: NavController,
    private alertController: AlertController) {}

  ngOnInit(): void {
    this.getData()
  }

  async getData() {
    let ul = await API.GetUsers();
    this.users = ul.Users;
    console.log(this.users);
  }

  onLogout() {
    localStorage.setItem("authToken", "");
    this.navcontroller.navigateRoot("/redux/login");
  }

  async presentAlertConfirmDelete(user: User) {
    const alert = await this.alertController.create({
      header: 'Confirm Deletion',
      message: 'Are you sure you want to delete the user '+user.Username+" ?",
      buttons: [
        {
          text: 'Cancel',
          role: 'cancel',
          cssClass: 'secondary',
          id: 'cancel-button',
        }, {
          text: 'Confirm',
          id: 'confirm-button',
          handler: () => {
            this.deleteUser(user);
          }
        }
      ]
    });

    

    await alert.present();
  }

  async deleteUser(user: User) {
    let ul = await API.DeleteUser(user.ID);
    this.users = ul.Users;
  }
}
