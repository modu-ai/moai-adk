import 'package:flutter/material.dart';

/// Example of responsive navigation that switches between
/// NavigationBar (bottom) and NavigationRail (side)
/// based on window width.
class ResponsiveNavigationExample extends StatelessWidget {
  const ResponsiveNavigationExample({super.key});

  @override
  Widget build(BuildContext context) {
    final width = MediaQuery.sizeOf(context).width;

    return Scaffold(
      body: width >= 600 ? _buildLargeLayout() : _buildSmallLayout(),
    );
  }

  /// Layout for small screens - bottom navigation
  Widget _buildSmallLayout() {
    return Scaffold(
      bottomNavigationBar: NavigationBar(
        destinations: const [
          NavigationDestination(icon: Icon(Icons.home), label: 'Home'),
          NavigationDestination(icon: Icon(Icons.search), label: 'Search'),
          NavigationDestination(icon: Icon(Icons.person), label: 'Profile'),
        ],
      ),
      body: const Center(child: Text('Small Layout')),
    );
  }

  /// Layout for large screens - side navigation rail
  Widget _buildLargeLayout() {
    return Scaffold(
      body: Row(
        children: [
          NavigationRail(
            destinations: const [
              NavigationRailDestination(icon: Icon(Icons.home), label: 'Home'),
              NavigationRailDestination(
                icon: Icon(Icons.search),
                label: 'Search',
              ),
              NavigationRailDestination(
                icon: Icon(Icons.person),
                label: 'Profile',
              ),
            ],
          ),
          const Expanded(child: Center(child: Text('Large Layout'))),
        ],
      ),
    );
  }
}
