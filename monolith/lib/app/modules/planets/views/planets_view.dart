import 'package:flutter/material.dart';

import 'package:get/get.dart';

import '../controllers/planets_controller.dart';

class PlanetsView extends StatelessWidget {
  final controller = Get.find<PlanetsController>();
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('Planets'),
        actions: [
          IconButton(
            onPressed: controller.create,
            icon: Icon(Icons.add_circle),
          ),
          IconButton(
            onPressed: controller.about,
            icon: Icon(
              Icons.favorite,
              color: Colors.redAccent,
            ),
          ),
        ],
      ),
      body: Container(
        padding: const EdgeInsets.symmetric(vertical: 15, horizontal: 10),
        child: GetBuilder(
          id: 'planets',
          init: controller,
          builder: (_) => ListView.builder(
            itemCount: controller.planets.length,
            itemBuilder: _itemBuilder,
          ),
        ),
      ),
    );
  }

  Widget _itemBuilder(BuildContext context, int index) {
    return GetBuilder(
      id: controller.planets[index].id,
      init: controller,
      builder: (_) {
        final planet = controller.planets[index];
        bool isOnline = controller.isOnline(planet);
        return Card(
          child: ListTile(
            title: FutureBuilder(
              future: controller.findServerName(planet.id),
              builder: (BuildContext context, AsyncSnapshot<String> snapshot) =>
                  Text(planet.viewName(snapshot.data)),
            ),
            subtitle: Text(
              planet.address,
              overflow: TextOverflow.ellipsis,
              maxLines: 1,
            ),
            trailing: PopupMenuButton(
              icon: Icon(Icons.more_vert),
              onSelected: (VoidCallback onTap) => onTap(),
              itemBuilder: (_) => [
                PopupMenuItem(
                  value: () => controller.view(planet),
                  child: ListTile(
                    leading: Icon(
                      Icons.videogame_asset,
                      color: Colors.purple,
                    ),
                    title: Text('Control'),
                  ),
                ),
                PopupMenuItem(
                  value: () => controller.edit(planet),
                  child: ListTile(
                    leading: Icon(
                      Icons.edit,
                      color: Colors.green,
                    ),
                    title: Text('Edit'),
                  ),
                ),
                PopupMenuItem(
                  value: isOnline ? () => controller.shutdown(planet) : null,
                  child: ListTile(
                    leading: Icon(
                      Icons.power_settings_new,
                      color: isOnline ? Colors.red : Colors.grey,
                    ),
                    title: Text('Disconnect'),
                  ),
                ),
                PopupMenuItem(
                  value: () => controller.delete(planet),
                  child: ListTile(
                    leading: Icon(
                      Icons.cancel,
                      color: Colors.red,
                    ),
                    title: Text('Delete'),
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }
}
