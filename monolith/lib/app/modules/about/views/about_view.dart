import 'package:flutter/material.dart';

import 'package:get/get.dart';
import 'package:monolith/app/share/ext.dart';
import 'package:package_info/package_info.dart';

import '../controllers/about_controller.dart';

class AboutView extends StatelessWidget {
  final packageInfo = Get.find<PackageInfo>();
  final controller = Get.find<AboutController>();

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          onPressed: Get.back,
          icon: Icon(Icons.chevron_left),
        ),
        title: Text('About'),
      ),
      body: ListView(
        children: [
          // controller.ab.setImage("assets/logo.png"),
          controller.ab.addDescription(
              " ${packageInfo.appName.upperWords()} \nRemote controll firmata. "),
          controller.ab.addWidget(Text('v${packageInfo.version}')),
          controller.ab.addGroup("Connect with us"),
          controller.ab.addEmail("xxx@gmail.com"),
          controller.ab.addFacebook("xxx"),
          controller.ab.addTwitter("xxx"),
          controller.ab.addYoutube("xxx"),
          controller.ab.addPlayStore("xxx"),
          controller.ab.addGithub("xxx"),
          controller.ab.addInstagram("xxx"),
          controller.ab.addWebsite("http://xxx"),
        ],
      ),
    );
  }
}
