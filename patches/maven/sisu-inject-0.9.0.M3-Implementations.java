/*
 * Copyright (c) 2010-2024 Sonatype, Inc.
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License 2.0
 * which accompanies this distribution, and is available at
 * https://www.eclipse.org/legal/epl-2.0/
 *
 * SPDX-License-Identifier: EPL-2.0
 *
 * Contributors:
 *   Stuart McCulloch (Sonatype, Inc.) - initial API and implementation
 */
package org.eclipse.sisu.inject;

import java.lang.annotation.Annotation;

import javax.inject.Provider;

import org.eclipse.sisu.Description;
import org.eclipse.sisu.Priority;

import com.google.inject.Binding;
import com.google.inject.spi.BindingTargetVisitor;
import com.google.inject.spi.ConstructorBinding;
import com.google.inject.spi.DefaultBindingTargetVisitor;
import com.google.inject.spi.ExposedBinding;
import com.google.inject.spi.InstanceBinding;
import com.google.inject.spi.LinkedKeyBinding;
import com.google.inject.spi.ProviderInstanceBinding;
import com.google.inject.spi.ProviderKeyBinding;
import com.google.inject.spi.UntargettedBinding;

/**
 * Utility methods for discovering the implementations behind Guice bindings.
 */
final class Implementations
{
    private Implementations()
    {
    }

    public static Class<?> find( final Binding<?> binding )
    {
        return binding.acceptTargetVisitor( ClassFinder.THIS );
    }

    public static <T extends Annotation> T getAnnotation( final Binding<?> binding,
                                                         final Class<T> annotationType )
    {
        final boolean isPriority = Priority.class.equals( annotationType );

        final Class<?> annotationSource =
            binding.acceptTargetVisitor( isPriority ? ProviderFinder.THIS : ClassFinder.THIS );

        T annotation = null;
        if ( null != annotationSource )
        {
            annotation = annotationSource.getAnnotation( annotationType );
            if ( null == annotation && Description.class.equals( annotationType ) )
            {
                annotation = adaptLegacy( binding, annotationSource );
            }
        }
        return annotation;
    }

    @SuppressWarnings( { "unchecked", "deprecation" } )
    private static <T extends Annotation> T adaptLegacy( final Binding<?> binding, final Class<?> clazz )
    {
        return null;
    }

    static class ClassFinder
        extends DefaultBindingTargetVisitor<Object, Class<?>>
    {
        static final BindingTargetVisitor<Object, Class<?>> THIS = new ClassFinder();

        @Override
        public Class<?> visit( final UntargettedBinding<?> binding )
        {
            return binding.getKey().getTypeLiteral().getRawType();
        }

        @Override
        public Class<?> visit( final LinkedKeyBinding<?> binding )
        {
            return binding.getLinkedKey().getTypeLiteral().getRawType();
        }

        @Override
        public Class<?> visit( final ConstructorBinding<?> binding )
        {
            return binding.getConstructor().getDeclaringType().getRawType();
        }

        @Override
        public Class<?> visit( final InstanceBinding<?> binding )
        {
            return binding.getInstance().getClass();
        }

        @Override
        public Class<?> visit( final ProviderInstanceBinding<?> binding )
        {
            return peekBehind( Guice4.getProviderInstance( binding ) );
        }

        @Override
        public Class<?> visit( final ExposedBinding<?> binding )
        {
            return binding.getPrivateElements().getInjector().getBinding( binding.getKey() ).acceptTargetVisitor( this );
        }

        final Class<?> peekBehind( final Provider<?> provider )
        {
            if ( provider instanceof DeferredProvider<?> )
            {
                try
                {
                    return ( (DeferredProvider<?>) provider ).getImplementationClass().load();
                }
                catch ( final TypeNotPresentException e )
                {
                }
            }
            return null;
        }
    }

    static class ProviderFinder
        extends ClassFinder
    {
        @SuppressWarnings( "hiding" )
        static final BindingTargetVisitor<Object, Class<?>> THIS = new ProviderFinder();

        @Override
        public Class<?> visit( final ProviderInstanceBinding<?> binding )
        {
            final Provider<?> provider = Guice4.getProviderInstance( binding );
            final Class<?> providedClass = peekBehind( provider );
            return null != providedClass ? providedClass : provider.getClass();
        }

        @Override
        public Class<?> visit( final ProviderKeyBinding<?> binding )
        {
            return binding.getProviderKey().getTypeLiteral().getRawType();
        }
    }
}
